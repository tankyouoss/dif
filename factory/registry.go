package factory

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

type RegistryCredentials struct {
	ServerURL string
	Username string
	Secret string
}

type DockerConfigAuth struct {
	Auth string `json:"auth"`
}

type DockerConfig struct {
	Credentials map[string]*DockerConfigAuth `json:"auths"`
	DefaultCredHelper   string `json:"credsStore"`
	CredHelpers map[string]string `json:"credHelpers"`
}

type RegistryAuthResponse struct {
	Token string
	ExpiresIn int `json:"expires_in"`
	IssuedAt string `json:"issued_at"`
}

type RegistryResponse struct {
	StatusCode int
	Body []byte
}

// This function parses the Www-Authenticate header provided in the challenge
// It has the following format
// Bearer realm="https://gitlab.com/jwt/auth",service="container_registry",scope="repository:andrew18/container-test:pull"
func parseBearer(bearer []string) map[string]string {
	out := make(map[string]string)
	for _, b := range bearer {
		for _, s := range strings.Split(b, " ") {
			if s == "Bearer" {
				continue
			}
			for _, params := range strings.Split(s, ",") {
				fields := strings.Split(params, "=")
				key := fields[0]
				val := strings.Replace(fields[1], "\"", "", -1)
				out[key] = val
			}
		}
	}
	return out
}

func readDockerConfig() (*DockerConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("coudn't get docker creds helper name\n%w", err)
	}

	file, err := os.Open(path.Join(home, ".docker/config.json"))
	if err != nil {
		return nil, fmt.Errorf("coudn't get docker creds helper name\n%w", err)
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)
	var dockerConfig DockerConfig
	json.Unmarshal(byteValue, &dockerConfig)
	return &dockerConfig, nil
}

func getCredentialHelper(dockerConfig *DockerConfig, registry string) (string, error) {
	helper := dockerConfig.DefaultCredHelper
	if value, ok := dockerConfig.CredHelpers[registry]; ok {
		helper = value
	}
	credHelper := fmt.Sprintf("docker-credential-%s", helper)
	return credHelper, nil
}

func GetRegistryCredentials(registry string) (*RegistryCredentials, error) {
	dockerConfig, err := readDockerConfig()
	if err != nil {
		return nil, fmt.Errorf("coudn't read docker config file\n%w", err)
	}
	credentials := RegistryCredentials{}

	if dockerConfigCreds, ok := dockerConfig.Credentials[registry]; ok && dockerConfigCreds != nil && len(dockerConfigCreds.Auth) > 0 {
		decodedCred, err := base64.StdEncoding.DecodeString(dockerConfigCreds.Auth)
		if err != nil {
			return nil, fmt.Errorf("coudn't read docker config credentials\n%w", err)
		}
		credComponents := strings.Split(string(decodedCred), ":")
		if len(credComponents) != 2 {
			return nil, fmt.Errorf("coudn't read docker config credentials. Missing username or password component for basic auth.\n")
		}
		credentials.Username = credComponents[0]
		credentials.Secret = credComponents[1]
		credentials.ServerURL = registry
	} else {
		credHelper, err := getCredentialHelper(dockerConfig, registry)
		if err != nil {
			return nil, fmt.Errorf("coudn't get docker registry %s credentials\n%w", registry, err)
		}

		execPath, err := exec.LookPath(credHelper)
		if err != nil {
			return nil, fmt.Errorf("coudn't get docker registry %s credentials\n%w", registry, err)
		}

		input := strings.NewReader(registry)
		cmd := &exec.Cmd {
			Path: execPath,
			Args: []string{ execPath, "get" },
			Stdin: input,
			//Stdout: os.Stdout,
			//Stderr: os.Stderr,
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("error running docker credentials helper.\n%w\n%s", err, output)
		}

		err = json.Unmarshal(output, &credentials)
		if err != nil {
			return nil, err
		}
	}

	return &credentials, nil
}

func getRegistryToken(params map[string]string, credentials *RegistryCredentials) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", params["realm"], nil)
	req.SetBasicAuth(credentials.Username, credentials.Secret)

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("coudn't get registry token\n%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("could't authenticate to %s. Got status code %d", params["realm"], resp.StatusCode)
	}

	authResponse := RegistryAuthResponse{}
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	if err != nil {
		return "", fmt.Errorf("coudn't read registry authentication response\n%w", err)
	}

	return authResponse.Token, nil
}

func registryGet(url string, credentials *RegistryCredentials) (*RegistryResponse, error) {
	client := &http.Client{}

	// First call with authentication
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coudn't query registry at url %s \n%w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("coudn't query registry at url %s \n%w", url, err)
		}

		response := RegistryResponse{
			StatusCode: resp.StatusCode,
			Body: body,
		}

		return &response, nil
	}

	// Get token
	authParams := parseBearer(resp.Header["Www-Authenticate"])
	token, err := getRegistryToken(authParams, credentials)
	if err != nil {
		return nil, fmt.Errorf("coudn't gettoken for registry at %s \n%w", url, err)
	}

	// Call with authentication token
	authReq, err := http.NewRequest("GET", url, nil)
	authReq.Header.Add("Authorization", "Bearer " + token)
	authResp, err := client.Do(authReq)
	if err != nil {
		return nil, err
	}
	defer authResp.Body.Close()

	body, err := ioutil.ReadAll(authResp.Body)
	if err != nil {
		return nil, fmt.Errorf("coudn't read registry response at url %s \n%w", url, err)
	}

	response := RegistryResponse{
		StatusCode: authResp.StatusCode,
		Body: body,
	}

	return &response, nil
}

func ImageExists(manifest Manifest) (bool, error) {
	creds, err := GetRegistryCredentials(manifest.Registry)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", manifest.Registry, manifest.Name, manifest.Tag)
	resp, err := registryGet(url, creds)
	if err != nil {
		return false, fmt.Errorf("coudn't query registry at url %s \n%w", url, err)
	}

	if resp.StatusCode != 404 && resp.StatusCode != 200 {
		return false, fmt.Errorf("coudn't query registry at url %s \n%s", url, string(resp.Body))
	}

	if resp.StatusCode == 404 {
		return false, nil
	}

	return true, nil
}
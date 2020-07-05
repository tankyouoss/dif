package factory

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"sort"
	"strings"
)

type Manifest struct {
	Registry string
	Name string
	Tag string
	AdditionalTags []string `yaml:"additionalTags"`
}

func deduplicate(array []string) []string {
	if len(array) <= 0 {
		return make([]string, 0)
	}

	input := make([]string, len(array))
	copy(input, array)
	sort.Strings(input)
	j := 0
	for i := 1; i < len(input); i++ {
		if input[j] == input[i] {
			continue
		}
		j++
		input[j] = input[i]
	}
	return input[:j+1]
}

func GitChangedFolders(repoPath string, currentSha string, previousSha string) ([]string, error){
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("invalid repository %s: %w", repoPath, err)
	}

	if len(currentSha) <= 0 {
		headRef, err := r.Head()
		if err != nil {
			return nil, fmt.Errorf("head reference not found: %w", err)
		}
		currentSha = headRef.Hash().String()
	}

	currentCommit, err := r.CommitObject(plumbing.NewHash(currentSha))
	if err != nil {
		return nil, fmt.Errorf("current commit (%s) not found: %w", currentSha, err)
	}

	previousCommit, err := r.CommitObject(plumbing.NewHash(previousSha))
	if err != nil {
		return nil, fmt.Errorf("previous commit (%s) not found: %w", previousSha, err)
	}

	patch, err := previousCommit.Patch(currentCommit)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the patch between (%s) and (%s) : %w", currentSha, previousSha, err)
	}

	var folders []string
	for _, file := range patch.Stats() {
		pathComponents := strings.Split(file.Name, "/")
		if len(pathComponents) > 1 {
			folders = append(folders, pathComponents[0])
		}
	}

	uniqFolders := deduplicate(folders)
	return uniqFolders, nil
}

func GitGetCurrentSha1(repoPath string) (string, error){
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid repository %s: %w", repoPath, err)
	}

	headRef, err := r.Head()
	if err != nil {
		return "", fmt.Errorf("head reference not found: %w", err)
	}
	currentSha := headRef.Hash().String()

	return currentSha, nil
}

func ReadManifest(repoPath string, directory string) (Manifest, error) {
	var manifest Manifest

	filepath := path.Join(repoPath, directory, "manifest.yml")
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return manifest, fmt.Errorf("couldn't read manifest file at (%s) : %w", filepath, err)
	}

	err = yaml.Unmarshal(yamlFile, &manifest)
	if err != nil {
		return manifest, fmt.Errorf("couldn't unmarshal manifest file at (%s) : %w", filepath, err)
		return manifest, err
	}

	return manifest, nil
}

func ImageName(manifest Manifest) string {
	return fmt.Sprintf("%s/%s:%s", manifest.Registry, manifest.Name, manifest.Tag)
}

func AdditionalImageNames(manifest Manifest) []string {
	names := make([]string, len(manifest.AdditionalTags))
	for idx, tag := range manifest.AdditionalTags {
		names[idx] = fmt.Sprintf("%s/%s:%s", manifest.Registry, manifest.Name, tag)
	}
	return names
}
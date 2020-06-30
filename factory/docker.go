package factory

import (
	"fmt"
	"io"
	"os/exec"
)

func Build(folderPath string, manifest Manifest, output io.Writer) (string, error) {
	execPath, err := exec.LookPath("docker")
	if err != nil {
		return "", err
	}

	imageName :=  ImageName(manifest)
	cmd := &exec.Cmd {
		Path: execPath,
		Args: []string{ execPath, "build", "-t", imageName, folderPath},
		Stdout: output,
		Stderr: output,
	}

	if err := cmd.Run() ; err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("coudn't build image %s. Exited with code %d", imageName, exitError.ExitCode())
		}
	}

	return imageName, nil
}

func PushImage(image string, output io.Writer) error {
	execPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	cmd := &exec.Cmd {
		Path: execPath,
		Args: []string{ execPath, "push", image},
		Stdout: output,
		Stderr: output,
	}

	if err := cmd.Run() ; err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("coudn't push image %s. Exited with code %d", image, exitError.ExitCode())
		}
	}

	return nil
}

func TagImage(image string, tag string, output io.Writer) error {
	execPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	cmd := &exec.Cmd {
		Path: execPath,
		Args: []string{ execPath, "tag", image, tag},
		Stdout: output,
		Stderr: output,
	}

	if err := cmd.Run() ; err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("coudn't tag image %s to %s. Exited with code %d", image, tag, exitError.ExitCode())
		}
	}

	return nil
}

func Push(image string, additionalTags []string, output io.Writer) error {
	err := PushImage(image, output)
	if err != nil {
		return err
	}

	for _, tag := range additionalTags {
		err := TagImage(image, tag, output)
		if err != nil {
			return err
		}

		err = PushImage(tag, output)
		if err != nil {
			return err
		}
	}

	return nil
}
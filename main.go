package main

import (
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/tankyouoss/dif/factory"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path"
)

const (
	done    = "\xE2\x9C\x94"
	failed  = "\xE2\x9D\x8C"
)

func build(repoPath string, previousSha1 string, currentSha1 string) {
	// Get diff
	fmt.Println(Yellow("Checking for images to build\n"))
	folders, err := factory.GitChangedFolders(repoPath, currentSha1, previousSha1)
	if err != nil {
		fmt.Println(Sprintf(Red("%s%v"), failed, err))
		os.Exit(1)
	}

	fmt.Println(Yellow("Found changes for"))
	for _, folder := range folders {
		fmt.Println(Sprintf(Yellow("  - %s"), folder))
	}
	fmt.Println("")

	for _, folder := range folders {
		manifest, err := factory.ReadManifest(repoPath, folder)
		if err != nil {
			fmt.Println(Sprintf(Red("%s%v"), failed, err))
			os.Exit(1)
		}

		imageName := factory.ImageName(manifest)
		fmt.Println(Sprintf(Yellow("Working on image %s in folder %s"), imageName, folder))

		// Check if it's a new tag
		imageExists, err := factory.ImageExists(manifest)
		if err != nil {
			fmt.Println(Sprintf(Red("%s%v"), failed, err))
			os.Exit(1)
		}

		if imageExists {
			fmt.Println(Sprintf(Red("%sYou probably forgot to update manifest tag. %s already exists."), failed, imageName))
			os.Exit(1)
		}

		fmt.Println(Sprintf(Green("    %sImage doesn't exists"), done))
		fmt.Println(Yellow("    Building image"))

		folderPath := path.Join(repoPath, folder)
		imageName, err = factory.Build(folderPath, manifest, os.Stdout)
		if err != nil {
			fmt.Println(Sprintf(Red("%s%v"), failed, err))
			os.Exit(1)
		}

		fmt.Println(Sprintf(Green("    %sImage successfully built\n"), done))
	}
}

func push(repoPath string, previousSha1 string, currentSha1 string) {
	// Get diff
	fmt.Println(Yellow("Checking for images to push\n"))
	folders, err := factory.GitChangedFolders(repoPath, currentSha1, previousSha1)
	if err != nil {
		fmt.Println(Sprintf(Red("%s%v"), failed, err))
		os.Exit(1)
	}

	fmt.Println(Yellow("Found changes for"))
	for _, folder := range folders {
		fmt.Println(Sprintf(Yellow("  - %s"), folder))
	}
	fmt.Println("")

	for _, folder := range folders {
		manifest, err := factory.ReadManifest(repoPath, folder)
		if err != nil {
			fmt.Println(Sprintf(Red("%s%v"), failed, err))
			os.Exit(1)
		}

		imageName := factory.ImageName(manifest)
		fmt.Println(Sprintf(Yellow("Working on image %s in folder %s"), imageName, folder))

		fmt.Println(Yellow("    Building image"))

		folderPath := path.Join(repoPath, folder)
		imageName, err = factory.Build(folderPath, manifest, os.Stdout)
		if err != nil {
			fmt.Println(Sprintf(Red("%s%v"), failed, err))
			os.Exit(1)
		}

		fmt.Println(Sprintf(Green("    %sImage successfully built"), done))

		fmt.Println(Yellow("    Pushing image"))
		additionalTags := factory.AdditionalImageNames(manifest)
		fmt.Printf("Additional tags : %v\n", additionalTags)
		fmt.Printf("manifest : %+v\n", manifest)
		err = factory.Push(imageName, additionalTags, os.Stdout)
		if err != nil {
			fmt.Println(Sprintf(Red("%s%v"), failed, err))
			os.Exit(1)
		}

		fmt.Println(Sprintf(Green("    %sImage successfully pushed\n"), done))
	}
}

func main() {
	app := &cli.App{
		Name: "DIF Docker Image Factory",
		Usage: "Build and push updated docker images",
		Version: "v0.0.1",
		Commands: []*cli.Command{
			{
				Name:    "build",
				Usage:   "Try to build updated docker images",
				ArgsUsage:   "previous_commit_sha1 [current_commit_sha1]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "repositoryPath", Aliases: []string{"rp"}},
				},
				Action:  func(c *cli.Context) error {
					if c.NArg() < 1 {
						fmt.Println(Sprintf(Red("%sPrevious commit sha1 is required"), failed))
						os.Exit(1)
					}

					repoPath := c.String("repositoryPath")
					if len(repoPath) <= 0 {
						repoPath = "./"
					}

					previousSha := c.Args().Get(0)
					currentSha := ""
					if c.NArg() > 1 {
						currentSha = c.Args().Get(1)
					}

					build(repoPath, previousSha, currentSha)
					//push("./tmp", "a0e38cba783dc9e9660bfdcaa9b12cb1cc6380a2", "")
					return nil
				},
			},
			{
				Name:    "push",
				Usage:   "Push updated docker images",
				ArgsUsage:   "previous_commit_sha1 [current_commit_sha1]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "repositoryPath", Aliases: []string{"rp"}},
				},
				Action:  func(c *cli.Context) error {
					if c.NArg() < 1 {
						fmt.Println(Sprintf(Red("%sPrevious commit sha1 is required"), failed))
						os.Exit(1)
					}

					repoPath := c.String("repositoryPath")
					if len(repoPath) <= 0 {
						repoPath = "./"
					}

					previousSha := c.Args().Get(0)
					currentSha := ""
					if c.NArg() > 1 {
						currentSha = c.Args().Get(1)
					}

					push(repoPath, previousSha, currentSha)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}


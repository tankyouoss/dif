# Docker image factory

This is still under heavy development

This is a tool to create a docker image factory based on git repository

It detects git changes to check which images needs to be updated.

It also checks that the version in the manifest file has been updated to ensure pushed are uniq.

## Usage

Each folder in the git repository need to be a docker image and should contain a manifest and dockerfile.
Example of manifest and folder architecture can be found in example folder.

Each build image has a label ``GitCommit`` with the repository current git commit sha1

## Build
To try ot build images and check version, run :
```
dif build --rp GIT_REPO_PATH PREVIOUS_COMMIT_SHA1 [CURRENT_COMMIT_SHA1]
```

For example : ``run run build --rp example 8750b695e9a9335211b491fad39cdeaf0d837843``

## Push
To push images run :
To try ot build images and check version, run :
```
dif push --rp GIT_REPO_PATH PREVIOUS_COMMIT_SHA1 [CURRENT_COMMIT_SHA1]
```

For example : ``dif push --rp example 8750b695e9a9335211b491fad39cdeaf0d837843``

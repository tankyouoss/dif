# Docker image factory

This is still under heavy development

This is a tool to create a docker image factory based on git repository

It detects git changes to check which images needs to be updated.

It also checks that the version in the manifest file has been updated to ensure pushed are uniq.

## Usage

Each folder in the git repository need to be a docker image and should contain a manifest and dockerfile.
Example of manifest and folder architecture can be found in example folder.

## Build
To try ot build images and check version, run :
`` dif build --rp GIT_REPO_PATH PREVIOUS_COMMIT_SHA1 [CURRENT_COMMIT_SHA1]``
For Example : ``run run build --rp example 8750b695e9a9335211b491fad39cdeaf0d837843``

To push images run :
To try ot build images and check version, run :
`` run run push --rp GIT_REPO_PATH PREVIOUS_COMMIT_SHA1 [CURRENT_COMMIT_SHA1]``
For Example : ``dif push --rp example 8750b695e9a9335211b491fad39cdeaf0d837843``

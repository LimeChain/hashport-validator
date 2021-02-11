# Release

To instantiate a new docker image release, you will need to publish a new [github](https://github.com/LimeChain/hedera-eth-bridge-validator/releases/new) release,
where the image tag will be specified tag and `latest` will become the new image. 
The image itself will be based on the content of the branch selected.
For example, publishing a release with tag `v0.0.1` will push the Docker image with tag `v0.0.1`, equivalent
to `gcr.io/hedera-eth-bridge-test/hedera-eth-bridge-validator:v0.0.1`.
###########################
# STEP 1 build the small image
# 
# NOTE: Instead of a 2 step approach
# we are using here only one single step
# more details can be found e.g.
# https://dev.to/stack-labs/introduction-to-taskfile-a-makefile-alternative-h92
# https://github.com/jeremyhuiskamp/golang-docker-scratch/blob/main/Dockerfile
############################
FROM scratch

# Copy our static executable.
COPY docker/bin/masteringsoundtouch-linux /
COPY docker/config.docker.toml /config.toml

# Run the hello binary.
ENTRYPOINT ["/masteringsoundtouch-linux"]
CMD ["-c", "/config.toml"]
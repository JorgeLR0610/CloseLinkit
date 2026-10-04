# ADR 0011: Use Distroless for the API Runtime Image

## Status

Accepted

## Context

CloseLinkit runs its HTTP API as a Go executable built from `server/` using a multi-stage Docker build. The current `server/Dockerfile` compiles the application with `CGO_ENABLED=0`, `GOOS=linux`, and `GOARCH=amd64`, then copies the resulting executable into `alpine:3.24.1`.

The runtime image currently creates an `appuser` account and runs the process as that user. However, the API does not require a general-purpose Linux environment to start: its entry point is a compiled Go binary, configuration is provided through environment variables, and database access is performed over the container network. The OpenAPI specification is embedded into the application, and the current entry point does not require shell utilities at runtime.

The project is moving toward deployment on an EC2 instance, initially with local development and testing retained through Docker Compose. A production runtime image should therefore minimize unnecessary packages and privilege while keeping the container build and operation straightforward.

The API currently logs through Go's `log/slog` package using a `TextHandler` that writes to `stderr`. Docker captures standard output and standard error independently of whether the runtime image contains a shell, so changing the base image does not require a different application logging implementation.

The primary options considered are:

* **Alpine:** a small Linux distribution with a shell and package manager. These tools can help with interactive debugging but are not needed by the API at runtime.
* **Distroless:** provides a minimal runtime image without a shell or package manager, while retaining commonly needed runtime files.
* **Scratch:** an empty image. It can run a suitable static executable, but any required system files must be copied or otherwise provided explicitly.

## Decision

Use `gcr.io/distroless/static-debian13:nonroot` as the API's final runtime image.

The Dockerfile will:

1. Keep a Go builder stage and compile the API with `CGO_ENABLED=0`, targeting Linux and the deployment architecture.
2. Copy only the compiled executable into the runtime image, at the absolute path `/server`.
3. Run the executable with `ENTRYPOINT ["/server"]`.
4. Rely on the `nonroot` image variant's configured unprivileged user instead of creating a user with `adduser` in the final stage.
5. Omit `WORKDIR` in the runtime stage because the executable is copied to an absolute path and the process does not need a working directory to start. If the application later depends on relative runtime file paths, this must be revisited.

The application will continue writing logs to `stdout` and/or `stderr`, and Docker's logging configuration will remain responsible for capturing them. Local development can continue using `docker logs`; centralized logging on EC2 can be introduced separately without coupling the application to a specific logging backend.

## Consequences

* **Positive:**

  * **Reduced runtime contents:** the final image does not include a shell, package manager, or general-purpose Alpine userland that the API does not need.
  * **Smaller attack surface:** fewer runtime components are present to maintain or potentially exploit.
  * **Non-root by default:** the selected image variant runs the API as an unprivileged user without custom user-creation commands.
  * **Separation of build and runtime:** the Go toolchain and build dependencies remain in the builder stage and are not copied into the final image.
  * **Logging remains unchanged:** `docker logs` continues to work because Docker captures the process's standard output and standard error independently of the image's shell availability.
  * **Straightforward runtime contract:** the container starts a single executable at `/server`.

* **Negative:**

  * **No interactive shell or package manager:** tools such as `sh`, `apk`, `curl`, and `ping` are unavailable inside the final image. Troubleshooting must use application logs, host-side tools, Docker inspection, or a separate debug image/container.
  * **Runtime assumptions must be explicit:** any future dependency on shared libraries, CA certificates, timezone data, configuration files, or other operating-system files must be checked when introducing it. The selected Distroless variant includes common CA and timezone data, but this must not be treated as a substitute for validating the application's actual requirements.
  * **Architecture must match deployment:** the builder's `GOARCH` setting must match the EC2 instance architecture; an `amd64` binary will not run natively on an `arm64` instance.
  * **Base-image updates still require maintenance:** the Distroless base image and Go builder image must be updated and rebuilt as part of normal dependency and security maintenance.
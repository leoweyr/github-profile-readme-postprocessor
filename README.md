# GitHub Profile Readme Postprocessor

A tool that leverages GitHub's built-in features to automatically update your GitHub profile readme file based on your situation and display needs, making the resume of your open-source journey on GitHub more direct, pure, and clear.

## 🏗️ Build

You can build the Docker image with optional build arguments to configure the Alpine mirror and Go proxy.

Please replace the `$AlpineMirror` and `$GoProxy` variables with their actual values:

```bash
docker build -t github-profile-postprocessor --build-arg ALPINE_MIRROR=$AlpineMirror GOPROXY=$GoProxy
```

| Argument        | Description                                      |
|-----------------|--------------------------------------------------|
| `ALPINE_MIRROR` | Alpine Linux package mirror domain.              |
| `GOPROXY`       | Go module proxy URL.                             |

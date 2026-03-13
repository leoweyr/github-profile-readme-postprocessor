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

## 🐙 GitHub Actions Usage

You can use this tool as a GitHub Action to automatically update your profile README.

> [!IMPORTANT]
>
> **Template-First**: Always edit your template file. The action overwrites `README.md` with rendered content.
> **Security**: All `endpoint` values MUST end with `/markdown`.
> **Permissions**: Ensure `permissions: contents: write` is set.

1. **Create Template**

   Create a template markdown file (e.g., `README.template.md`) in your repository. Insert the following anchors where you want content:

   ```markdown
   <!-- YOUR_ANCHOR -->
   ```

2. **Configure Workflow**

   For full endpoint usage and parameter details, please refer to [api/openapi.yaml](api/openapi.yaml).

   ```yaml
   name: Update Profile README
   
   on:
     schedule:
       - cron: '0 0 * * *'  # Run daily at midnight.
     workflow_dispatch:
   
   jobs:
     update-readme:
       runs-on: ubuntu-latest
       permissions:
         contents: write  # Essential for pushing changes.
       steps:
         - uses: actions/checkout@v4
         
         - name: Update Profile Readme
           uses: leoweyr/github-profile-readme-postprocessor@main
           with:
             github_token: ${{ secrets.GITHUB_TOKEN }}
             readme_template_path: 'README.template.md'  # Your template file.
             tasks: |
               [
                 {
                   "anchor": "<!-- SUPPORT_INFO -->",
                   "endpoint": "/v1/support/markdown",
                   "params": {}
                 },
                 {
                   "anchor": "<!-- CONTRIBUTED_REPOS -->",
                   "endpoint": "/v1/contributed-repositories/markdown",
                   "params": {
                     "limit_count": 5,
                     "include_commits": true,
                     "title": "### 📦 Project"
                   }
                 }
               ]
         
         - name: Commit and Push
           run: |
             git config --global user.name 'github-actions[bot]'
             git config --global user.email 'github-actions[bot]@users.noreply.github.com'
             git add README.md
             git commit -m "docs: update profile rstats" || exit 0
             git push
   ```

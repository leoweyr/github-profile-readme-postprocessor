![github-profile-readme-postprocessor](https://socialify.git.ci/leoweyr/github-profile-readme-postprocessor/image?description=1&font=KoHo&logo=https%3A%2F%2Fraw.githubusercontent.com%2Fleoweyr%2Fgithub-profile-readme-postprocessor%2Frefs%2Fheads%2Fdevelop%2Fassets%2Ficon.svg&name=1&owner=1&pattern=Formal+Invitation&theme=Light)

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

![Usage](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fabacus.jasoncameron.dev%2Fget%2Fleoweyr%2Fgithub-profile-readme-postprocessor-usage&query=%24.value&label=Usage&color=blue&suffix=%20times)
![Used by Developers](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/leoweyr/64244ccf3d52ec6458cb56c652a41c8d/raw/github-profile-readme-postprocessor-github-actions-used-by-stats.json)

You can use this tool as a GitHub Action to automatically update your profile README.

> [!IMPORTANT]
>
> **Template-First**: Always edit your template file. The action overwrites `README.md` with rendered content.
> **Security**: All `endpoint` values MUST end with `/markdown`.
> **Permissions**: Ensure `permissions: contents: write` is set.

| Argument | Description | Required |
| :--- | :--- | :--- |
| `github_token` | GitHub Token for API access. | **Yes** |
| `readme_template_path` | Path to the README template file. | No |
| `tasks` | JSON list of tasks defining anchors, endpoints, and params. | No |
| `sort_latest_activity_blocks` | Whether to sort `LATEST_ACTIVITY` blocks by timestamp (descending). | No |

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
             sort_latest_activity_blocks: true
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

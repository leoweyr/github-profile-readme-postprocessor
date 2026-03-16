# Changelog

All notable changes to this project will be documented in this file.

# [0.3.1](https://github.com/leoweyr/github-profile-readme-postprocessor/compare/v0.3.0...v0.3.1) (2026-03-16)
### Bug Fixes

* **contributed-repositories-markdown:** parse adaptive show recent activity stats param ([b30f0ed](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/b30f0eda91036a43a590c1470e57ee62e975c9b7)) [@leoweyr](https://github.com/leoweyr)
* **contributed-repositories-markdown:** reliably discover both public and private repositories, bypassing search API limitations ([9bf83cd](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/9bf83cd4abb2fd25244a0df84520f3619b58bc8b)) [@leoweyr](https://github.com/leoweyr)



# [0.3.0](https://github.com/leoweyr/github-profile-readme-postprocessor/compare/v0.2.1...v0.3.0) (2026-03-16)
### Bug Fixes

* **contributed-repositories-markdown:** truncate commit message to first line in latest activity ([995ab2b](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/995ab2be0d7a1dd79b26724deb34f0c37e6453fd)) [@leoweyr](https://github.com/leoweyr)
* **contributed-repositories-markdown:** resolve premature limit truncation before filtering repositories ([aac619c](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/aac619cb0dd4f727cf1360ed8c16a8c1e612609e)) [@leoweyr](https://github.com/leoweyr)
* **trend-topics-markdown:** correct endpoint registration ([1382be4](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/1382be404e81cf6d868e43cf4e9fa3336ea3cfbc)) [@leoweyr](https://github.com/leoweyr)
* **contributed-repositories-markdown:** prevent 500 error on public activity fetch failure ([7e97bea](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/7e97bea8f96f4b6b39be2792191f7f99d76deebe)) [@leoweyr](https://github.com/leoweyr)


### Features

* **contributed-repositories-markdown:** add support for private repositories ([464f073](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/464f073cf03c0476cf511c8d3f357d4b378cefde)) [@leoweyr](https://github.com/leoweyr)
* **trend-topics-markdown:** add endpoint with time-decay weighting algorithm ([3f438b6](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/3f438b62bc89af064990c6af764c5fa32c13188a)) [@leoweyr](https://github.com/leoweyr)


### Performance

* **contributed-repositories-markdown:** optimize contributed repositories fetching with ladder strategy ([c2fcf12](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/c2fcf1203cc80c1ff9f59c9c533d783af958932a)) [@leoweyr](https://github.com/leoweyr)



# [0.2.1](https://github.com/leoweyr/github-profile-readme-postprocessor/compare/v0.2.0...v0.2.1) (2026-03-14)
### Bug Fixes

* **github-action-app:** convert boolean parameters to lowercase strings for Go server compatibility ([074fc85](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/074fc85ae45c1c261fd85afb88c58079a6a1a0b5)) [@leoweyr](https://github.com/leoweyr)


### Features

* **github-action-app:** add debug logging for fetched content ([12b6c84](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/12b6c8408fc33eda2ac739941c2ffd8cc5848e57)) [@leoweyr](https://github.com/leoweyr)
* **github-action-app:** add include request URL in debug log output ([2859341](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/2859341ec9ea8e0e72611e1fe5efced3f9d824c0)) [@leoweyr](https://github.com/leoweyr)



# [0.2.0](https://github.com/leoweyr/github-profile-readme-postprocessor/compare/v0.1.0...v0.2.0) (2026-03-14)
### Bug Fixes

* **contributed-repositories-markdown:** include user issues in activity statistics ([8b0b5b4](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/8b0b5b4832e3560f6b91ecf130a471ac77af889f)) [@leoweyr](https://github.com/leoweyr)


### Features

* **contributed-repositories-markdown:** add activity timestamp to markdown response ([a97b2ad](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/a97b2adc7c298238124997e73ff28656cdea9cec)) [@leoweyr](https://github.com/leoweyr)
* **github-action-app:** support sorting activity blocks by timestamp ([d3de7d1](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/d3de7d1006e4e0ed9f7aaf5bbefbdaa66fdcb5f4)) [@leoweyr](https://github.com/leoweyr)
* **contributed-repositories-markdown:** add recent activity stats ([22b499c](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/22b499c2460d398ff6c6dfbd49887c02a9b6ea48)) [@leoweyr](https://github.com/leoweyr)
* **contributed-repositories-markdown:** add adaptive recent activity stats ([c6a755a](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/c6a755a55136cb24ce5e4564582f499de6c8b03b)) [@leoweyr](https://github.com/leoweyr)
* **contributed-repositories-markdown:** add show latest activity ([1db7fac](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/1db7facdf18526ef533898017fd9c5eda56c4698)) [@leoweyr](https://github.com/leoweyr)


### Documentation

* **readme:** add banner ([70b5c7a](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/70b5c7a0b8ad4e0ef7be9b447bb3f1ab020c9ff0)) [@leoweyr](https://github.com/leoweyr)


### Miscellaneous Tasks

* add icon ([8a80d54](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/8a80d5412851dd6ff6846916191545a6384ebb40)) [@leoweyr](https://github.com/leoweyr)
* add branding configuration for marketplace release ([eacb847](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/eacb8474c900ca39418828b920a123d5f22e7327)) [@leoweyr](https://github.com/leoweyr)



# [0.1.0] (2026-03-13)
### Bug Fixes

* **terraform:** inject missing APP_LISTEN_PORT env var ([ad1d33d](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/ad1d33dbfe4e5722ef834adb80dba1122e051a6d)) [@leoweyr](https://github.com/leoweyr)
* increase FC function timeout to 60s ([fde5155](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/fde5155892e34c2223b0fbca2606a3768649ce07)) [@leoweyr](https://github.com/leoweyr)
* increase server timeouts to prevent premature connection closure ([2311760](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/231176058b8cdbbb85b4fce2e2186c9fcc95dc16)) [@leoweyr](https://github.com/leoweyr)
* **contributed-repositories:** correctly split comma-separated filter parameters ([b376b51](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/b376b5150dca4cfc84ba24d163e53211ee82472e)) [@leoweyr](https://github.com/leoweyr)
* **github-action-app:** set correct working directory for Go server startup in composite action ([4ea63d8](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/4ea63d85b57c7de65fa59c43c8e7e4bbb5cac244)) [@leoweyr](https://github.com/leoweyr)
* **github-action-app:** update PYTHONPATH to include scripts directory and adjust module imports for direct execution ([36c7a00](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/36c7a00a2770a6633a55e505cd79c131ee7bbb95)) [@leoweyr](https://github.com/leoweyr)


### Features

* **fetcher:** add retrieve user starred repositories within a date range ([f4a5e87](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/f4a5e87469c2b37d7e318fc254bf5159e122668b)) [@leoweyr](https://github.com/leoweyr)
* **domain:** add repository ([1d0553f](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/1d0553f0e4d3e3c8bc347c8e8a4aed6ceb6713db)) [@leoweyr](https://github.com/leoweyr)
* **fetcher:** add retrieve user watched repositories within 90 days ([af14bc8](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/af14bc8ac9f887bfaee307fe2502f25f09f3ccc2)) [@leoweyr](https://github.com/leoweyr)
* **fetcher:** add retrieve user forked repositories within a date range ([f8b0801](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/f8b0801187f2a018f3333296ae7584d94d202d64)) [@leoweyr](https://github.com/leoweyr)
* **fetcher:** add retrieve user pushed repositories within a date range ([bdda3e5](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/bdda3e590f696b4dcaff9e8aec044d632e32d7fc)) [@leoweyr](https://github.com/leoweyr)
* **fetcher:** add retrieve user pushed commits within a date range ([c72d5be](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/c72d5bee07542d3a0f562a556f0f1484c950fae0)) [@leoweyr](https://github.com/leoweyr)
* **fetcher:** add retrieve user issue activities within a date range ([a0cc6cd](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/a0cc6cd5f05f0f3cbba2e746f86dfc33c061e6ba)) [@leoweyr](https://github.com/leoweyr)
* **domain:** add topics to repository and update fetchers to populate it ([2a61689](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/2a61689e49fdd83222fd903ef894293b17754449)) [@leoweyr](https://github.com/leoweyr)
* **filter:** add repository filtering by name and topic ([2d074d4](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/2d074d488de96c7679cb0b64721bf59bf383b525)) [@leoweyr](https://github.com/leoweyr)
* **api:** add definition for fetching contributed repositories ([1334580](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/1334580ba369a2ab45c58a90d9ee5d724bd53b3c)) [@leoweyr](https://github.com/leoweyr)
* add `/contributed-repositories` endpoints stub ([07c3073](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/07c3073d385030eb3dedec1f94d68f08d414fbc3)) [@leoweyr](https://github.com/leoweyr)
* **fetcher:** add retrieve user pull request activities within a date range ([8140ae2](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/8140ae20497fcff195aaa9c1c86ba8672f3e09e7)) [@leoweyr](https://github.com/leoweyr)
* implement `/contributed-repositories` endpoint ([9f87fe4](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/9f87fe4db214bf3c0158147f816ee759acf8ab85)) [@leoweyr](https://github.com/leoweyr)
* **http:** add global panic recovery middleware ([82c90d2](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/82c90d2e51ac8b89ae2a78a4b1a7618b7601a20b)) [@leoweyr](https://github.com/leoweyr)
* **api:** add definition for fetching contributed repositories markdown ([db52ce1](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/db52ce10b4ce8bbba780d6088446fc6f9dff7685)) [@leoweyr](https://github.com/leoweyr)
* implement `/contributed-repositories/markdown` endpoint ([cbb925e](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/cbb925ebaef8bdd47e8b225850aebcb624ba385b)) [@leoweyr](https://github.com/leoweyr)
* support array-based filtering for repository names and topics ([a2dfce4](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/a2dfce4b5fa740b02d2a593a560124286f3dd643)) [@leoweyr](https://github.com/leoweyr)
* **contributed-repositories:** combine repositories name and topic filters with OR semantics ([45d201d](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/45d201dc59bf83d8d55df9bf36c9cb59fa67adfc)) [@leoweyr](https://github.com/leoweyr)
* **support:** add endpoint to return project info in markdown ([b1f75c9](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/b1f75c9f39fe67bb04d7ce8cb9c79d27c3acd7b4)) [@leoweyr](https://github.com/leoweyr)
* package application as a reusable composite GitHub Action for automated profile updates ([dff431d](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/dff431d7147cf7a4f1f1640d5c3c7bb4f1b2aa00)) [@leoweyr](https://github.com/leoweyr)


### Performance

* boost FC memory to 512MB and timeout to 120s ([d69fb9c](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/d69fb9c2f02f8a45658eb41c00ce1fcd1a229830)) [@leoweyr](https://github.com/leoweyr)


### Refactor

* restructure entire project architecture as a Go project ([fcf3194](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/fcf3194187b1a3cc7e507ca768489f15149d43c9)) [@leoweyr](https://github.com/leoweyr)
* remove main application entry point ([dc63d03](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/dc63d03feeb9b09e01fa5868e891922ad74c51ee)) [@leoweyr](https://github.com/leoweyr)
* restructure project to align with Clean Architecture ([46b4dc1](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/46b4dc145002d2ce3f31f641d53834b3adf71cd6)) [@leoweyr](https://github.com/leoweyr)


### DevOps

* enable cloud native deployment to Alibaba Cloud FC ([985ec76](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/985ec76b5e2c1e6b1556a18fa9dbb0334580d6c3)) [@leoweyr](https://github.com/leoweyr)
* skip APISIX registration if service unreachable or credentials missing ([c900c3c](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/c900c3cedee6afcefa3d5420ce83ea747a7210be)) [@leoweyr](https://github.com/leoweyr)
* fix syntax error by removing unnecessary ${{ }} from conditional ([225c1b5](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/225c1b57ac945ab25f62170f8b95312f84c6f220)) [@leoweyr](https://github.com/leoweyr)
* rely on script logic for APISIX validation instead of workflow condition ([6e9e2c6](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/6e9e2c673510545e09f90215a8e95165ca7a132b)) [@leoweyr](https://github.com/leoweyr)
* correct AliCloud fc function config ([25e7c18](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/25e7c18858bdb87db9d9e9395b8b3da62c325f8b)) [@leoweyr](https://github.com/leoweyr)
* fix terraform compatibility with AliCloud fc custom runtime ([3bd5d2f](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/3bd5d2f933b15843daee8935b9a4c21c7d12b126)) [@leoweyr](https://github.com/leoweyr)
* remove reserved env var FC_CUSTOM_LISTEN_PORT ([7f358c1](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/7f358c1f6974a900de33ce70dadb543cb5031248)) [@leoweyr](https://github.com/leoweyr)
* configure proxy-rewrite to strip routing prefix for upstream compatibility ([f284a24](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/f284a248dc502e18f0a27fb1ca5cf50c3b20384a)) [@leoweyr](https://github.com/leoweyr)
* add automated release workflows ([8e6fe86](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/8e6fe869c6588ebaa3785207c611207b4b149c5b)) [@leoweyr](https://github.com/leoweyr)


### Styling

* standardize GitHub API token environment variable naming ([ba96f20](https://github.com/leoweyr/github-profile-readme-postprocessor/commit/ba96f20fcff62c379805eae16931407221c1549d)) [@leoweyr](https://github.com/leoweyr)



<!-- Generated by git-cliff. -->

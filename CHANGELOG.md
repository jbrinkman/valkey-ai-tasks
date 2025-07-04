## [1.4.3](https://github.com/jbrinkman/valkey-ai-tasks/compare/v1.4.2...v1.4.3) (2025-07-04)

### Bug Fixes

* add support for linux/arm64 to the base image ([8aa3115](https://github.com/jbrinkman/valkey-ai-tasks/commit/8aa3115ab01d090dfcc427284fa6ca60ab92c6ad))

## [1.4.2](https://github.com/jbrinkman/valkey-ai-tasks/compare/v1.4.1...v1.4.2) (2025-07-04)

### Bug Fixes

* adjust base valkey image for dockerfile ([cd6db61](https://github.com/jbrinkman/valkey-ai-tasks/commit/cd6db612b4b8f16715e64102a9c3a5ebfd32683d))

## [1.4.1](https://github.com/jbrinkman/valkey-ai-tasks/compare/v1.4.0...v1.4.1) (2025-07-02)



# [1.4.0](https://github.com/jbrinkman/valkey-ai-tasks/compare/v1.3.0...v1.4.0) (2025-07-02)


### Features

* target amd64 platform specification for container builds ([3e8c41b](https://github.com/jbrinkman/valkey-ai-tasks/commit/3e8c41b3eddef579e633fa674fe654e4ab4bf1c3))



# [1.3.0](https://github.com/jbrinkman/valkey-ai-tasks/compare/v1.2.0...v1.3.0) (2025-07-02)


### Features

* add multi-platform container build support for amd64/arm64/windows ([d598eb1](https://github.com/jbrinkman/valkey-ai-tasks/commit/d598eb15ee5b0c3188d396a59d8874422c6cbae5))



# [1.2.0](https://github.com/jbrinkman/valkey-ai-tasks/compare/v1.1.2...v1.2.0) (2025-07-02)


### Features

* implement MCP resources and tools for plan/task management ([b23f91e](https://github.com/jbrinkman/valkey-ai-tasks/commit/b23f91ed15b5583c897dda30cba19733e8fd633b))



## [1.1.2](https://github.com/jbrinkman/valkey-ai-tasks/compare/v1.1.0...v1.1.2) (2025-06-30)


### Bug Fixes

* add linting to the project and fix existing lint errors ([c7b4efd](https://github.com/jbrinkman/valkey-ai-tasks/commit/c7b4efdc6ed115a008c72ef539d13f6c5f11ace1))



# [1.1.0](https://github.com/jbrinkman/valkey-ai-tasks/compare/v1.0.0...v1.1.0) (2025-06-28)


### Bug Fixes

* add required GitHub permissions for release workflows and actions ([ed6535d](https://github.com/jbrinkman/valkey-ai-tasks/commit/ed6535d5f13016e2e8f0dec560613afbd9af2592))
* replace GITHUB_TOKEN with MY_TOKEN to allow workflow events to trigger next workflow ([bc926b0](https://github.com/jbrinkman/valkey-ai-tasks/commit/bc926b05deaafd5cb6b85e7d0aa5b5511b878d42))


### Features

* add manual workflow trigger with version input for container publishing ([497140e](https://github.com/jbrinkman/valkey-ai-tasks/commit/497140e86ece71dc2644a24d35bfee94813caee6))



# [1.0.0](https://github.com/jbrinkman/valkey-ai-tasks/compare/da77265dcb897e7830450c1ee3b0a1d313b21659...v1.0.0) (2025-06-27)


### Bug Fixes

* add --fail-on-empty=false to commitlint command ([7d3d387](https://github.com/jbrinkman/valkey-ai-tasks/commit/7d3d38737129835c678bc00e65aa9064317475e8))
* address PR review comments ([f1b7c69](https://github.com/jbrinkman/valkey-ai-tasks/commit/f1b7c699226cb664bbe2c50ca5d19cb52f8ed8b1))
* improve semantic release logic and fix build-and-push job condition ([3a9dd77](https://github.com/jbrinkman/valkey-ai-tasks/commit/3a9dd77f6793a4490ed3979758896b920726dbf3))
* increase server startup delay in transport test and remove unused transport type constants ([52ad98a](https://github.com/jbrinkman/valkey-ai-tasks/commit/52ad98a06a93315b11b272272d01a4a870b433c2))
* iterate over map keys directly instead of using range with unused value ([37b7aca](https://github.com/jbrinkman/valkey-ai-tasks/commit/37b7aca4d35ec2b8bdf8fd0d6369c1033b98f3b0))
* prevent test suite from using port 6379 to avoid conflicts with dev instances ([059862d](https://github.com/jbrinkman/valkey-ai-tasks/commit/059862d5aa592b240439b782ab14ece764e9d772))
* remove fail-on-empty flag from commitlint validation ([5db064a](https://github.com/jbrinkman/valkey-ai-tasks/commit/5db064a9a62ce5c5415203e6dbb324ec7783eaa6))
* replace --extends with --config in semantic-release command ([df451b9](https://github.com/jbrinkman/valkey-ai-tasks/commit/df451b90ee7d73e8b38b379406b9313dca86c422))
* update .github/workflows/commit-lint.yml ([0196c41](https://github.com/jbrinkman/valkey-ai-tasks/commit/0196c417a0200bcfc8e6ebea0f4107b539390db2))
* update .github/workflows/container-build.yml ([fb723ee](https://github.com/jbrinkman/valkey-ai-tasks/commit/fb723eeeca46a8540d8740a0c078d1687ea68f70))
* update MCP server URLs to use /sse endpoint instead of /mcp ([5df56ac](https://github.com/jbrinkman/valkey-ai-tasks/commit/5df56ac858b9ddfad1c5abc1cbef2cbcb6c06404))
* update transport config to use streamable-http instead of streamable_http ([6e66b95](https://github.com/jbrinkman/valkey-ai-tasks/commit/6e66b95d7d18ef1f333f1bcd5d60ad1babb144fb))
* update Valkey port from 16379 to 6379 in docker-compose configuration ([875c596](https://github.com/jbrinkman/valkey-ai-tasks/commit/875c59652aa235cfd9adf779b93522af412bf3d0))


### Features

* add application association to projects with filtering capabilities ([eb8f2bf](https://github.com/jbrinkman/valkey-ai-tasks/commit/eb8f2bf350a4cdf36f168a7a6976f14da847f1eb))
* add bulk task creation endpoint and repository method for efficient multi-task operations ([c006826](https://github.com/jbrinkman/valkey-ai-tasks/commit/c0068263211896a497ea6c2854e78c5d9c8e3379))
* add bulk task creation with integration tests and helper functions ([6ce5084](https://github.com/jbrinkman/valkey-ai-tasks/commit/6ce5084d83c75d8c624cfe6bed49bb5ec658fc1b))
* add changelog generation for manual releases ([93f11bc](https://github.com/jbrinkman/valkey-ai-tasks/commit/93f11bccef0aa718192cf6f5065f3bb3e49746da))
* add ListByPlanAndStatus method to filter tasks by plan and status ([a1145aa](https://github.com/jbrinkman/valkey-ai-tasks/commit/a1145aa0a71d33ff2154beb088cb018e05262f77))
* add markdown notes field to Project and Task models ([1f6ff46](https://github.com/jbrinkman/valkey-ai-tasks/commit/1f6ff46ad6b28c24603a640c6ec5cf6e42d824fb))
* add markdown validation, sanitization and formatting utilities for notes ([6d3af98](https://github.com/jbrinkman/valkey-ai-tasks/commit/6d3af98614a55545bb80396687a31c311be60e6f))
* add Markdown-formatted notes support for projects and tasks with API endpoints ([3ac3894](https://github.com/jbrinkman/valkey-ai-tasks/commit/3ac3894a581158b6acd8cd346401d64f7b58a7fe))
* Add MCP and Valkey in a single container. Remove compose files ([d2d428d](https://github.com/jbrinkman/valkey-ai-tasks/commit/d2d428dc9757bdfabedba6beb0c3efb68181d3fa))
* add MCP server implementation with Valkey persistence for task management ([da77265](https://github.com/jbrinkman/valkey-ai-tasks/commit/da77265dcb897e7830450c1ee3b0a1d313b21659))
* add notes support to plans and tasks with legacy project cleanup ([df85de2](https://github.com/jbrinkman/valkey-ai-tasks/commit/df85de257cfa74bbb285ed5cb232cfc84137bf84))
* add orphaned tasks functionality and integration tests ([4f61394](https://github.com/jbrinkman/valkey-ai-tasks/commit/4f613946443cf0cca5306a33dae3ef55a301807d))
* add repository interfaces and initial test coverage setup ([3831c61](https://github.com/jbrinkman/valkey-ai-tasks/commit/3831c6161952aebcecc1a392a44d2f672ae4e9f2))
* add SSE support and implement MCP-Go server integration ([32ba91b](https://github.com/jbrinkman/valkey-ai-tasks/commit/32ba91b4b99d75bba607bcc5b5ba7b1f64222331))
* add STDIO transport support for legacy AI tool integration ([ecaca2a](https://github.com/jbrinkman/valkey-ai-tasks/commit/ecaca2a596dbc4fbb21aa365cd3554a88fd6fad9))
* add support for Markdown notes in plans and tasks with CRUD operations ([2c040a9](https://github.com/jbrinkman/valkey-ai-tasks/commit/2c040a966cdb7e21711354e902f749e89b0028fe))
* add support for streamable HTTP transport alongside SSE ([d5b10f1](https://github.com/jbrinkman/valkey-ai-tasks/commit/d5b10f18faae7b3edb1231f9bd1a0d1f3fe1e0c4))
* add test infrastructure with Makefile, helpers, and integration test setup ([3bde443](https://github.com/jbrinkman/valkey-ai-tasks/commit/3bde443afbadacd4e46f29c6821034fb4614cc87))
* add test utilities for Valkey container and test data management ([f8e0d1e](https://github.com/jbrinkman/valkey-ai-tasks/commit/f8e0d1e6309053c26cb17facca7579f32690e1f8))
* add workflow_dispatch support for manual releases ([4a44693](https://github.com/jbrinkman/valkey-ai-tasks/commit/4a44693b3efd87ec3b87bfc5e95192b89368ca11))
* create initial plan for adding STDIO transport support to MCP server ([75a68b3](https://github.com/jbrinkman/valkey-ai-tasks/commit/75a68b37f03bd31e8fdc2501c341fce5e8a19574))
* implement CI/CD workflow for container image builds and releases ([8fdb287](https://github.com/jbrinkman/valkey-ai-tasks/commit/8fdb287b370a1d09ba0283f382bcee953cd83eca)), closes [#18](https://github.com/jbrinkman/valkey-ai-tasks/issues/18)
* switch to SSE server and update tool descriptions for feature planning focus ([e24d8af](https://github.com/jbrinkman/valkey-ai-tasks/commit/e24d8afb04480952f18ad6019467ba34dd8de366))

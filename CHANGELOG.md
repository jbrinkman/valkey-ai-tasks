## 1.0.0 (2025-06-27)

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

### Documentation

* add example agent prompts for Valkey MCP task management system ([768139e](https://github.com/jbrinkman/valkey-ai-tasks/commit/768139e1da5c5dea0fbb31cbc939ecf67ff2fe71))
* add MCP configuration guide and agent usage examples to README ([b04aae0](https://github.com/jbrinkman/valkey-ai-tasks/commit/b04aae0fdc77bc5ae5bdcd633834b66e5a4eb198))

### Code Refactoring

* consolidate integration tests into test suites ([051de85](https://github.com/jbrinkman/valkey-ai-tasks/commit/051de85080afe2ab383794323c95002c8b9baa6e))
* consolidate notes tools into server_mcp.go and add repository tests ([dc41685](https://github.com/jbrinkman/valkey-ai-tasks/commit/dc41685d6b1449c9e76ab02d28a2fb4152031b38))
* consolidate plan repository tests into single test suite ([61dee00](https://github.com/jbrinkman/valkey-ai-tasks/commit/61dee009c9d5c85738c87221060a44415e70b944))
* Extract plan status validation into shared helper function ([a485fcb](https://github.com/jbrinkman/valkey-ai-tasks/commit/a485fcb362c0ab14ef4ec161c34a07c6c5a509c2))
* remove mock repositories and unused Valkey default port constant ([5eef8ca](https://github.com/jbrinkman/valkey-ai-tasks/commit/5eef8ca78e35f13e2f6902a2c3b7f3247bf875dd))
* remove unused project repository and example test files ([5f7a876](https://github.com/jbrinkman/valkey-ai-tasks/commit/5f7a876f9c42c57585a24afa0e10005c203b94fa))
* rename project terminology to plan terminology throughout codebase ([988390e](https://github.com/jbrinkman/valkey-ai-tasks/commit/988390e200b8ffe5f46dca0882c493e5ced25c54))
* rename project to plan throughout MCP server implementation ([d932f65](https://github.com/jbrinkman/valkey-ai-tasks/commit/d932f65ca6ed3e70a3da3775388d265ae409771d))
* reorganize project structure by moving Go files to root directory ([7af5216](https://github.com/jbrinkman/valkey-ai-tasks/commit/7af5216a3a30a7cc2d8960e4ddf5306024a6a533))
* update storage layer to use new Valkey Glide v2 API ([b764791](https://github.com/jbrinkman/valkey-ai-tasks/commit/b7647912a8069245a93f1397ac218fedf7de0fc0))

### Tests

* add edge case and concurrency tests for plan repository ([8461d68](https://github.com/jbrinkman/valkey-ai-tasks/commit/8461d684688eb0d563902870ccf18b73e63b9da1))
* add integration tests for plan and task notes management ([02ce521](https://github.com/jbrinkman/valkey-ai-tasks/commit/02ce521ff80c6bf91df5f554028bef9b1203a225))
* add integration tests for plan repository and edge cases ([9024954](https://github.com/jbrinkman/valkey-ai-tasks/commit/90249549a889f5e9c54bf8b2ee3d9617bae03a1b))
* add integration tests for task repository and fix reordering logic ([344afb6](https://github.com/jbrinkman/valkey-ai-tasks/commit/344afb6c30486134666487c97e41fb78b3c9c1d3))
* add MCP server integration tests with random port allocation ([1fd895e](https://github.com/jbrinkman/valkey-ai-tasks/commit/1fd895e8a830c73e0c99558f372e010a4209d86e))
* update MCP server connection test to not use default endpoints and improve error handling ([dc49ee2](https://github.com/jbrinkman/valkey-ai-tasks/commit/dc49ee29c3fc74d9ea0d5b94c780c61bfba9405c))

### Continuous Integration

* configure semantic-release to use conventional commits preset ([1101272](https://github.com/jbrinkman/valkey-ai-tasks/commit/1101272c228bc5d4d204e4978a4c618fcf4d0215))
* restructure workflow with PR validation and separate release paths ([21bb0b6](https://github.com/jbrinkman/valkey-ai-tasks/commit/21bb0b617725e972383290668b5c777db70e4931))

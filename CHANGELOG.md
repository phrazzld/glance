## [1.1.4](https://github.com/phrazzld/glance/compare/v1.1.3...v1.1.4) (2026-02-23)


### Bug Fixes

* **prompt:** use relative paths in LLM prompt instead of absolute ([#69](https://github.com/phrazzld/glance/issues/69)) ([ceb9052](https://github.com/phrazzld/glance/commit/ceb9052a4c708b3d4fc653749fa349dc571a8c3a)), closes [#51](https://github.com/phrazzld/glance/issues/51)

## [1.1.3](https://github.com/phrazzld/glance/compare/v1.1.2...v1.1.3) (2026-02-22)


### Bug Fixes

* skip LLM for empty directories, write minimal stub instead ([#63](https://github.com/phrazzld/glance/issues/63)) ([40547bf](https://github.com/phrazzld/glance/commit/40547bfb6b4cbe864bb32a7526c1417a84c619ea)), closes [#61](https://github.com/phrazzld/glance/issues/61)

## [1.1.2](https://github.com/phrazzld/glance/compare/v1.1.1...v1.1.2) (2026-02-22)


### Bug Fixes

* **prompt:** tighten default template to prevent hallucination ([#62](https://github.com/phrazzld/glance/issues/62)) ([2453cf2](https://github.com/phrazzld/glance/commit/2453cf2b57102a0a03da66f87377b4562224cb20)), closes [#246](https://github.com/phrazzld/glance/issues/246)

## [1.1.1](https://github.com/phrazzld/glance/compare/v1.1.0...v1.1.1) (2026-02-21)


### Bug Fixes

* **output:** rename glance.md to .glance.md to avoid build system conflicts ([#58](https://github.com/phrazzld/glance/issues/58)) ([f956d59](https://github.com/phrazzld/glance/commit/f956d591a90e20265e4208526e8de4fc06f4b390)), closes [#50](https://github.com/phrazzld/glance/issues/50) [/github.com/phrazzld/glance/pull/58#discussion_r2836447758](https://github.com//github.com/phrazzld/glance/pull/58/issues/discussion_r2836447758)

# [1.1.0](https://github.com/phrazzld/glance/compare/v1.0.0...v1.1.0) (2026-02-21)


### Features

* **config:** default to current directory when no target specified ([c7b1ec0](https://github.com/phrazzld/glance/commit/c7b1ec0c6d8842d4ee98887ba530f8264cbecc09))

# 1.0.0 (2026-02-21)


### Bug Fixes

* accept STOP as valid finish reason from Gemini API ([e2b29b7](https://github.com/phrazzld/glance/commit/e2b29b7e10aeb6123adf8a24b41a2de1f8a75254))
* apply pre-commit formatting to Dependabot integration documentation ([e6bdcd5](https://github.com/phrazzld/glance/commit/e6bdcd59bb2b19bc23c812fc87cd16e1677e5a85))
* avoid non-text files, retry on failed attempts ([57f66b5](https://github.com/phrazzld/glance/commit/57f66b5464a659e159a26f91c51d7d584bf562b2))
* configure no-commit-to-branch to only run during commit stage ([98d4b19](https://github.com/phrazzld/glance/commit/98d4b192d3a947e851b28ddd6b00d9b00a6b773e))
* **gitignore:** use it lol ([c9323c6](https://github.com/phrazzld/glance/commit/c9323c697d8ffb1eb5cd0389210d9b16350023ec))
* implement parent directory regeneration tracking ([25a146d](https://github.com/phrazzld/glance/commit/25a146d833cd298db6aa86b5797a22881191eb6a))
* improve custom hook to correctly block direct commits ([eb3235a](https://github.com/phrazzld/glance/commit/eb3235a00b7f728e0a3c8a03dced807db3f366d2))
* include error message in non-verbose ReportError (T214) ([8b1da70](https://github.com/phrazzld/glance/commit/8b1da70e3d711545d5cd9b132b42f0795c78556e))
* linter warnings, error handling, and failing tests ([97c6988](https://github.com/phrazzld/glance/commit/97c698845b57461c3535d57c9ddb7b773d59d7ad))
* mark T021 as complete - missing test infrastructure restored ([2514455](https://github.com/phrazzld/glance/commit/2514455b1276503ef166eac7d162a03779047ee4))
* pin govulncheck version to v1.1.3 across all workflows ([ea3a9c4](https://github.com/phrazzld/glance/commit/ea3a9c4874eebf3ee8a54f4524622bab9c44358c))
* **recursion:** only pull glance files from subdirectories ([8fb0233](https://github.com/phrazzld/glance/commit/8fb0233701bae8f9a69fb18bd1969a53be475068))
* reduce timeout test duration from 1s to 100ms for reliable timeout behavior ([da1d892](https://github.com/phrazzld/glance/commit/da1d8921c9e0f534adca64e4c0d689137feee1fe))
* remove empty baseDir escape hatch in path validation (T204) ([43d9bdf](https://github.com/phrazzld/glance/commit/43d9bdf60ef821b864407ec3e88054565977f02d))
* remove flawed glance_file_missing check in logging ([2f45cb9](https://github.com/phrazzld/glance/commit/2f45cb9cbaa16a181d92a58ddd46c06644650b45))
* remove inverted condition blocking Windows path test ([e2be001](https://github.com/phrazzld/glance/commit/e2be001d3e565c6d50f84b8cd03921dd3cf94a2d))
* remove trailing whitespace from network failure test ([5712721](https://github.com/phrazzld/glance/commit/57127211ec2265ffc1fbd3e54a210b4564e65b3e))
* remove trailing whitespace from security scanning documentation ([b9f3327](https://github.com/phrazzld/glance/commit/b9f3327bae2460780728545bf987cf376bc77e20))
* replace no-commit-to-branch with custom hook that allows merges ([e360c35](https://github.com/phrazzld/glance/commit/e360c357f50ce30fa0ff9855e8bf6422e4b119f4))
* resolve CI failures and consolidate Go version testing ([c6e1b85](https://github.com/phrazzld/glance/commit/c6e1b85a0494e6c1da738e20974a5c47167cd4d4))
* resolve CI pipeline failures and improve test reliability ([5887813](https://github.com/phrazzld/glance/commit/588781317e4f83a56ab0c49b1079b4001bb19286))
* resolve vulnerability test string pattern matching for govulncheck v1.1.3 ([52cd5fd](https://github.com/phrazzld/glance/commit/52cd5fd485e492567d897d4f86ef21f61a08e11e))
* secure prompt template loading against path traversal (T201) ([3907399](https://github.com/phrazzld/glance/commit/3907399cbd7380aa61802e9c3d79565f4d18b898))
* temporarily skip failing tests to unblock CI pipeline ([647a4e2](https://github.com/phrazzld/glance/commit/647a4e2d92fec87f2b2d6c7785b2d90fa833e5bf))
* update clean project test to accept different govulncheck output formats ([57849f6](https://github.com/phrazzld/glance/commit/57849f6bf67b4f1c3aa9cf12819279d5cb3b5675))
* update golangci-lint config schema compatibility for v2.1.2 ([93817bf](https://github.com/phrazzld/glance/commit/93817bf0ed682903fa42a916df98d07a6dde86b2))
* update golangci-lint configuration for v2.1.2 compatibility ([54aaacf](https://github.com/phrazzld/glance/commit/54aaacf9e7118aeac5af43e9fc7ed7208e713b18))
* update golangci-lint GitHub Action to v7 ([81b3a7f](https://github.com/phrazzld/glance/commit/81b3a7fea7bd639e6dfa9dbd6cc0b35b9050d3b1))
* update pre-commit workflow to use relocated requirements.txt ([7c396c8](https://github.com/phrazzld/glance/commit/7c396c82922a895b02741a6c6f5878223965c69f))
* update test expectations for simplified govulncheck behavior ([7a9fb53](https://github.com/phrazzld/glance/commit/7a9fb5339e4a1573c230a1e2005dec21f5f06be6))
* use prompt.txt file ([86150db](https://github.com/phrazzld/glance/commit/86150db334a6b4d956c3751915c12d0af5e1bfbf))


### Features

* add DefaultFileMode constant (T051) ([3d157c9](https://github.com/phrazzld/glance/commit/3d157c929a71100ce9efca931bc179af7db6ddc5))
* add environment diagnostics to CI workflows for better debugging ([1b46e67](https://github.com/phrazzld/glance/commit/1b46e67f0b0a0c429acfee2e31d34af60e9243e0))
* add file length check to pre-commit hooks ([44aee3a](https://github.com/phrazzld/glance/commit/44aee3abcace52d80d8220dd938c9af920b6f69b))
* add GitHub Actions summary for vulnerability scans ([7218a2e](https://github.com/phrazzld/glance/commit/7218a2e76a4e7be9c6bd69335f4e47bb4dd50ca6))
* add post-commit hook to run glance asynchronously ([cb068f6](https://github.com/phrazzld/glance/commit/cb068f641d6e75444aa36057ba7851b26d6d97b6))
* add PromptTemplate to ServiceOptions (T047) ([96ad7ca](https://github.com/phrazzld/glance/commit/96ad7ca5263f3cb52e491f2614289093adbfb7bd))
* add test output suppression for progress bar ([8d6c43c](https://github.com/phrazzld/glance/commit/8d6c43cd8d6dc8ba0554e6c4de51a1611658d48e))
* add WithPromptTemplate option function (T048) ([18698f6](https://github.com/phrazzld/glance/commit/18698f6cc9eea698351a01c08dafb3ad3d7d6b27))
* complete emergency override protocol and artifact management ([8ffdbd9](https://github.com/phrazzld/glance/commit/8ffdbd93b0d757134cc4f0ffcacc9dbd18584e97))
* complete T016 vulnerability database caching investigation and optimization ([e3452d9](https://github.com/phrazzld/glance/commit/e3452d95e0f24085fd40d86045bfef42c166cc7a)), closes [#57150](https://github.com/phrazzld/glance/issues/57150)
* **config:** create Config package and struct ([e79c0e6](https://github.com/phrazzld/glance/commit/e79c0e6afc0c15600ca22654fc68d6f397139a33))
* **config:** implement LoadConfig function ([027f5f1](https://github.com/phrazzld/glance/commit/027f5f1c74acd7c30fdb0f021705eed5d3b6c3ae))
* configure severity-based failure thresholds for vulnerability scanning ([27b1f88](https://github.com/phrazzld/glance/commit/27b1f881ebee1d98d7d25be00684d93ff60a0ebf))
* enhance logging with emojis and descriptive messages ([17562da](https://github.com/phrazzld/glance/commit/17562da831b6f4bc464e182261323d3b6d7b870d))
* **errors:** define custom error types for improved error handling ([ea13a4e](https://github.com/phrazzld/glance/commit/ea13a4e7dba388a98d474587defaf7b130e2f984))
* **errors:** implement error wrapping with helper functions ([2ac6004](https://github.com/phrazzld/glance/commit/2ac600453def6b21096f29a55aaeacae168b37af))
* **filesystem:** centralize ignore logic in dedicated module ([38f2eac](https://github.com/phrazzld/glance/commit/38f2eacab608ff3fb01bd010a5b9d63574b563ce))
* **filesystem:** create reader module with file reading and text detection functions ([8e403f2](https://github.com/phrazzld/glance/commit/8e403f262941ca29e0e7202db95cf38d3c673fb2))
* **filesystem:** create scanner package with gitignore support ([9eb75cd](https://github.com/phrazzld/glance/commit/9eb75cdecbaeadcae7a879f378e56cccbaefc7f9))
* **filesystem:** create utilities module with file modification time helpers ([2be7386](https://github.com/phrazzld/glance/commit/2be73861114a3ac91d366f5a16364fafe24daf7d))
* generate structured vulnerability reports with correlation tracking ([b4b1eff](https://github.com/phrazzld/glance/commit/b4b1effb230dc2412e0df3929bfee1fd55fea6cb))
* implement comprehensive network failure handling for vulnerability scanning (T015) ([58f0c17](https://github.com/phrazzld/glance/commit/58f0c173d84e4399e748b45b520f437832197bd3))
* implement direct progress bar usage in glance.go ([df0771a](https://github.com/phrazzld/glance/commit/df0771a368e44002cda79927ea79f6b8f972a662))
* implement emergency override and enhanced security scanning ([d350b12](https://github.com/phrazzld/glance/commit/d350b12885b90656eb1c364ba42021d44ffeec32))
* implement fail-fast pipeline termination with enhanced error messaging ([f3c38a9](https://github.com/phrazzld/glance/commit/f3c38a901fbe1dc0d60a032088ab69436cb1037b))
* implement path validation for external inputs (T038) ([6f57170](https://github.com/phrazzld/glance/commit/6f57170f1e9cb5a4a21de184ffb380fa979f54b6))
* implement path validation utilities for T037 ([d81e77a](https://github.com/phrazzld/glance/commit/d81e77a950fc41a87aef66d748afca8105187d28))
* implement structured logging using logrus fields ([ea1dd77](https://github.com/phrazzld/glance/commit/ea1dd77449713e86b28d9f52d7b12bd5d659d45a))
* implement structured logging with correlation ID propagation ([2fb2921](https://github.com/phrazzld/glance/commit/2fb2921f58b8402a637a3cb6d9366405c1495257))
* improve glance.md prompt template ([81f32d5](https://github.com/phrazzld/glance/commit/81f32d51cb4263c42473b7f4a6b536042b1c7c99))
* integrate govulncheck vulnerability scanning into CI pipeline ([573f6f3](https://github.com/phrazzld/glance/commit/573f6f3ff4b80d4673f73b3b5bb3ece4c3f659f4)), closes [#20](https://github.com/phrazzld/glance/issues/20)
* integrate vulnerability scanning metrics with observability platforms ([dfcd8ff](https://github.com/phrazzld/glance/commit/dfcd8ff3610d6f173fa4f9455edeb703fdb8172e))
* **llm:** add robust retry and model/provider failover chain ([#46](https://github.com/phrazzld/glance/issues/46)) ([c4ec6d9](https://github.com/phrazzld/glance/commit/c4ec6d910286e56ac9360090f10caee2a0650ff3))
* **llm:** add streaming API support for the new GenAI client ([5cc54c5](https://github.com/phrazzld/glance/commit/5cc54c5e38dce2285413c93c6b949c69d736548a))
* **llm:** complete interface review (T009) ([08fd02a](https://github.com/phrazzld/glance/commit/08fd02a350f4c0673950f489948e6c734f672598))
* **llm:** complete validation of main application code (T012) ([6d24f82](https://github.com/phrazzld/glance/commit/6d24f82fedce9d2ca8d5b4e2640b5718be932ea7))
* **llm:** create prompt generation module ([d0db5d2](https://github.com/phrazzld/glance/commit/d0db5d29c7b54448e7236048422fd03501ecbbad))
* **llm:** define client interface with Gemini implementation ([ba332be](https://github.com/phrazzld/glance/commit/ba332be61317549929aa0d0a2f8146fd0dac6096))
* **llm:** enhance token counting logic for new GenAI package ([08c9b71](https://github.com/phrazzld/glance/commit/08c9b715694f939b7bfe31a5a79ece6ab6c611e6))
* **llm:** migrate to google.golang.org/genai package (T001, T002) ([2d55a9f](https://github.com/phrazzld/glance/commit/2d55a9fa97fc8582f69f70b58e2a7be36c645df4))
* **llm:** refactor client to use non-streaming API calls ([d0ec047](https://github.com/phrazzld/glance/commit/d0ec047816ce858152c69c5684d37e3b1174391e))
* **llm:** update functional options for google.golang.org/genai package ([21fcb40](https://github.com/phrazzld/glance/commit/21fcb4009d449b4b02c788e83f1be551bf2da871))
* **llm:** update mocks for llm.Client interface (T010) ([5e78d21](https://github.com/phrazzld/glance/commit/5e78d21d603477f7a9f08b6ca56f81ffa1dc7824))
* **llm:** update model selection logic to use new package (T003) ([f9daded](https://github.com/phrazzld/glance/commit/f9daded44e357d0b38c3265e856aeb799c5c2765))
* **llm:** update service layer to work with refactored client (T011) ([6322fb3](https://github.com/phrazzld/glance/commit/6322fb38fdef62347aeb1028e2027b8a099f522c))
* **llm:** verify all tests work with GenAI implementation (T015) ([42bda3b](https://github.com/phrazzld/glance/commit/42bda3b665e44851b51f13945f87813710f8327c))
* **llm:** verify unit tests compatibility with GenAI (T013, T014) ([2423d92](https://github.com/phrazzld/glance/commit/2423d92e30f1c175d98150242e644161bf8f1363))
* make logging level configurable via environment variable ([44668be](https://github.com/phrazzld/glance/commit/44668becff345617ab2a9832524b243c124e2cfa))
* migrate backlog to GitHub issues and clean up repository ([2ca9bc9](https://github.com/phrazzld/glance/commit/2ca9bc96e00aa361696c36226965dd257d4b4f39))
* regen glance files if files they cover have been modified ([10d4546](https://github.com/phrazzld/glance/commit/10d454696c9822289eb16b16ae732da4d718c013))
* **tests:** setup testing framework with testify ([178c1ae](https://github.com/phrazzld/glance/commit/178c1aefc46809fa0d5fae23f8e867e8cecaa53b))
* **tests:** update integration tests for google.golang.org/genai (T016) ([5fc14e6](https://github.com/phrazzld/glance/commit/5fc14e6ad4e0277e1fbb0003bd0e5c9e9b563882))
* **ui:** create UI feedback module for spinner and progress bar management ([de700ae](https://github.com/phrazzld/glance/commit/de700ae11dadf52566a79244fd9178888fbb85e4))
* update service to use stored template (T049) ([3facfa9](https://github.com/phrazzld/glance/commit/3facfa950345722410d0e14725dfda1342445c9d))

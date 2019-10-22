# Auth Server

Auth Service provides the authentication for the cuttle.ai platform

## Installation

```bash
go get -v github.com/cuttle-ai/auth-service
```

## Usage

Navigate into the project directory and run the following command

```bash
go run main.go
```

### Environment Variables

The environment varibals are loaded from vault. So if the following env are configured and the rest can be configured inside the vault.

*CUTTLE_AI_CONFIG_VAULT_ADDRESS* - Address at which vault is available. Eg. https://vault.cuttle.ai
*CUTTLE_AI_CONFIG_VAULT_TOKEN* - Vault authentication token
*CUTTLE_AI_CONFIG_VAULT_DEFAULT_PATH* - Path at which vault screts are stored. Eg. cuttle-ai-development

| Enivironment Variable                 | Description                                                                                     |
| ------------------------------------- | ----------------------------------------------------------------------------------------------- |
| **PORT**                              | Port on to which application server listens to. Default value is 8080                           |
| **RESPONSE_TIMEOUT**                  | Timeout for the server to write response. Default value is 100ms                                |
| **REQUEST_BODY_READ_TIMEOUT**         | Timeout for reading the request body send to the server. Default value is 20ms                  |
| **RESPONSE_BODY_WRITE_TIMEOUT**       | Timeout for writing the response body. Default value is 20ms                                    |
| **PRODUCTION**                        | Flag to denote whether the server is running in production. Default value is `false`            |
| **SKIP_VAULT**                        | Skip loading the configurations from vault server. Default value is `false`.                    |
| **IS_TEST**                           | Denoting the run is test. This will load the test configuration from vault                      |
| **MAX_REQUESTS**                      | Maximum no. of concurrent requests supported by the server. Default value is 1000               |
| **REQUEST_CLEAN_UP_CHECK**            | Time interval after which error request app context cleanup has to be done. Default value is 2m |
| **OAUTH2_GOOGLE_REDIRECT_URL**        | Google oauth redirect url                                                                       |
| **OAUTH2_GOOGLE_CLIENT_ID**           | Google oauth client id                                                                          |
| **OAUTH2_GOOGLE_CLIENT_SECRET**       | Google oauth client secret                                                                      |
| **OAUTH2_GOOGLE_USER_PROFILE_SCOPE**  | Google user profile scope                                                                       |
| **OAUTH2_GOOGLE_USER_EMAIL_SCOPE**    | Google user email scope                                                                         |
| **OAUTH2_GOOGLE_USER_INFO_URL**       | Google exchange url which gives google user info                                                |
| **OAUTH2_GOOGLE_USER_INFO_NAME**      | Key storing the name key in the google user info                                                |
| **OAUTH2_GOOGLE_USER_INFO_EMAIL**     | Key storing the email key in the google user info                                               |
| **FRONTEND_URL**                      | URL for accessing the frontend                                                                  |
| **DISCOVERY_URL**                     | URL of the discovery service consul                                                             |
| **DISCOVERY_TOKEN**                   | Access token for accessing the discovery service consul                                         |

## Author

Melvin Davis<hi@melvindavis.me>

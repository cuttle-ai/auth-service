# Auth Server

Auth Service provides the authentication for the cuttle.ai platform

## Prerequisite

You would require the following to be installed in your system

- [node](https://nodejs.org/en/)
- [go](https://golang.org/)

## Installation

Add the following variables to your .bashrc or .zshrc in your home directory

```
export VAULT_ROOT_KEY='get-the-vault-root-token-from-development-team'
export CUTTLE_AI_CONFIG_VAULT_TOKEN=$VAULT_ROOT_KEY
export CUTTLE_AI_CONFIG_VAULT_ADDRESS='https://vault.cuttle.ai'
export CUTTLE_AI_CONFIG_VAULT_DEFAULT_PATH='cuttle-ai-development'
```

```bash
git clone https://github.com/cuttle-ai/auth-service
cd auth-service
sh setup.sh
cd ../brain-frontend
sudo npm install -g @angular/cli
npm i
```

## Usage

Navigate into the project directory and run the following command

```bash
cd ../brain-frontend && npm start
```

Open another terminal session in the project directory and run the following command

```bash
go run main.go
```

Now open the browser and navigate to [localhost:4200](http://localhost:4200). Authenticate yourself using the Google login offered by the platform.
Open Developer Tools(Browser) -> Application -> Cookies , Use the cookie value of `auth-token` for testing API

### Environment Variables

| Enivironment Variable                | Description                                                                                     |
| ------------------------------------ | ----------------------------------------------------------------------------------------------- |
| **PORT**                             | Port on to which application server listens to. Default value is 8080                           |
| **RESPONSE_TIMEOUT**                 | Timeout for the server to write response. Default value is 100ms                                |
| **REQUEST_BODY_READ_TIMEOUT**        | Timeout for reading the request body send to the server. Default value is 20ms                  |
| **RESPONSE_BODY_WRITE_TIMEOUT**      | Timeout for writing the response body. Default value is 20ms                                    |
| **PRODUCTION**                       | Flag to denote whether the server is running in production. Default value is `false`            |
| **SKIP_VAULT**                       | Skip loading the configurations from vault server. Default value is `false`.                    |
| **IS_TEST**                          | Denoting the run is test. This will load the test configuration from vault                      |
| **MAX_REQUESTS**                     | Maximum no. of concurrent requests supported by the server. Default value is 1000               |
| **REQUEST_CLEAN_UP_CHECK**           | Time interval after which error request app context cleanup has to be done. Default value is 2m |
| **OAUTH2_GOOGLE_REDIRECT_URL**       | Google oauth redirect url                                                                       |
| **OAUTH2_GOOGLE_CLIENT_ID**          | Google oauth client id                                                                          |
| **OAUTH2_GOOGLE_CLIENT_SECRET**      | Google oauth client secret                                                                      |
| **OAUTH2_GOOGLE_USER_PROFILE_SCOPE** | Google user profile scope                                                                       |
| **OAUTH2_GOOGLE_USER_EMAIL_SCOPE**   | Google user email scope                                                                         |
| **OAUTH2_GOOGLE_USER_INFO_URL**      | Google exchange url which gives google user info                                                |
| **OAUTH2_GOOGLE_USER_INFO_NAME**     | Key storing the name key in the google user info                                                |
| **OAUTH2_GOOGLE_USER_INFO_EMAIL**    | Key storing the email key in the google user info                                               |
| **FRONTEND_URL**                     | URL for accessing the frontend                                                                  |
| **DISCOVERY_URL**                    | URL of the discovery service consul                                                             |
| **DISCOVERY_TOKEN**                  | Access token for accessing the discovery service consul                                         |

## Author

[Melvin Davis](mailto:melvinodsa@gmail.com)

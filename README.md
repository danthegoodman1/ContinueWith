# ContinueWith

Become an OAuth2 provider with any auth backend.

ContinueWith is a service that proxies the OAuth2 flow between your backend and other apps that want to use you as an oauth provider. It handles:

1. Registering apps
2. All authorization flows (client credentials, authorization code, device code)
3. Refresh and access token management
4. Scope management

All you need to implement is:

1. A pretty OAuth consent screen that matches your awesome site (we have a stellar guide to help you crush it quickly)
2. An API endpoint that lets
3. Define your scopes!
4. Any auth middleware you need to check an access token against ContinueWith and get back user info, scopes, etc.

<!-- TOC -->
* [ContinueWith](#continuewith)
  * [Admin API](#admin-api)
<!-- TOC -->

## User API

The user api is the endpoint that apps use to handle the oauth2 flow with you. It looks roughly like this:

(insert flow chart)

## Admin API

The admin api allows you to manage apps, scopes, and more.
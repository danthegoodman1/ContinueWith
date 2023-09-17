# ContinueWith

Become an OAuth2 provider with any auth backend.

ContinueWith is a service that proxies the OAuth2 flow between your backend and clients (apps that want to use you as an oauth provider). It handles:

1. Registering clients
2. All authorization flows (client credentials, authorization code, device code)
3. Refresh and access token management
4. Scope management

What you need to do is:

1. Make a pretty OAuth consent screen that matches your awesome site (we have a stellar guide to help you crush it quickly)
2. Make an API endpoint that we can forward your bearer token or session to when your users give consent, and you give us some user info
3. Tell us what scopes are available
4. Add any auth middleware you need to check the access token against ContinueWith and get back user info, scopes, etc.

<!-- TOC -->
* [ContinueWith](#continuewith)
  * [User API](#user-api)
  * [Admin API](#admin-api)
<!-- TOC -->

## User API

The user api is the endpoint that clients use to handle the oauth2 flow with you. It looks roughly like this:

(insert flow chart)

## Admin API

The admin api allows you to manage clients, scopes, and more.
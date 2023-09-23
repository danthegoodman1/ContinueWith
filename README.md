# ContinueWith

Become an OAuth2 provider with any auth backend.

ContinueWith is a service that proxies the OAuth2 flow between your backend and clients (apps that want to use you as an oauth provider). It handles:

1. Registering clients
2. All authorization flows (client credentials, authorization code, device code)
3. Refresh and access token management
4. Scope management

For example, maybe you use Firebase, Supabase, or Clerk for manage your users and want to allow other sites to login users and access resources from your platform. ContinueWith manages this oauth flow on top of your existing auth system.

Notion does this: They allow their users to log in with social providers like Google, and also allow other platforms to log in their users with Notion and access things like pages and databases through the Notion API.

<!-- TOC -->
* [ContinueWith](#continuewith)
  * [All you need to do:](#all-you-need-to-do)
  * [Examples](#examples)
  * [User API](#user-api)
  * [Admin API](#admin-api)
<!-- TOC -->

## All you need to do:

1. Make a pretty OAuth consent screen that matches your awesome site (we have a stellar guide to help you crush it quickly)
2. Make an API endpoint that we can forward your bearer token or session to when your users give consent, and you give us some user info (you hit this from your consent screen)
3. Define your scopes
4. Add any auth middleware you need to check the access token against ContinueWith and get back user info, scopes, etc. (you just hit our API endpoint to get a userid back from the access token)

## Examples

See the [example](example) provider and client to get a feel for what the behavior looks like, and how easy it is to implement as a provider.

## User API

The user api is the endpoint that clients use to handle the oauth2 flow with you. It looks roughly like this:

(insert flow chart)

## Admin API

The admin api allows you to check access tokens, manage clients, scopes, and more.
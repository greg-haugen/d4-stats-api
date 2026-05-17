# Platform Auth Setup

## Battle.net (`BNET_CLIENT_ID`, `BNET_CLIENT_SECRET`)

1. Go to https://develop.battle.net/ and sign in with your Blizzard account.
2. Click **Create Client** and fill in the application name and redirect URI (any value works for client-credentials flow).
3. Copy the **Client ID** and **Client Secret** from the client detail page.
4. Set `BNET_CLIENT_ID=<id>` and `BNET_CLIENT_SECRET=<secret>` in your environment.

The service uses the OAuth 2.0 client credentials flow — no user login required.

> **Note:** As of the time of writing, Blizzard has not published official public API endpoints for Diablo 4 player stats. The Battle.net client is implemented against the expected endpoint pattern (`/d4/profile/{username}`). Verify availability at https://develop.battle.net/documentation before deploying.

---

## Steam (`STEAM_API_KEY`)

1. Visit https://steamcommunity.com/dev/apikey while logged into your Steam account.
2. Enter a domain name (any value is accepted for server-side use).
3. Copy the generated API key.
4. Set `STEAM_API_KEY=<key>` in your environment.

The key is sent as the `key` query parameter on every Steam Web API request.

---

## PSN (`PSN_AUTH_TOKEN`)

PSN does not provide a server-to-server OAuth flow. You must supply an **npsso token** obtained from an authenticated PSN session.

1. Sign in to https://store.playstation.com in a browser.
2. Visit `https://ca.account.sony.com/api/v1/ssocookie` — the response JSON contains your `npsso` value.
3. Set `PSN_AUTH_TOKEN=<npsso>` in your environment.

**Important:** npsso tokens are short-lived (typically 60 days). When the token expires, the service returns `502 PSN_TOKEN_INVALID` to callers. Rotate the token by repeating the steps above and restarting the service.

---

## Xbox Live (`XBOX_CLIENT_ID`, `XBOX_CLIENT_SECRET`)

1. Go to https://portal.azure.com and sign in with a Microsoft account.
2. Navigate to **Azure Active Directory → App registrations → New registration**.
3. Under **Certificates & secrets**, create a new client secret and copy the value immediately.
4. Grant the application the required Xbox Live permissions (or add them under **API permissions**).
5. Set `XBOX_CLIENT_ID=<application-id>` and `XBOX_CLIENT_SECRET=<secret>` in your environment.

The service performs a two-step flow: Microsoft OAuth token → XSTS token exchange. Both tokens are cached in memory and refreshed automatically.

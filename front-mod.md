# Frontend Sign In And Sign Up Fix

## What was broken

Both sign up and sign in were failing from the frontend, but for two different browser-facing reasons:

1. `Sign up` could fail because the backend CORS policy only trusted one old hardcoded S3 website origin.
2. `Sign in` could appear to work at `/auth/login` but then fail on later authenticated requests because the frontend depended on cross-site cookies from a plain S3 static website setup.

Since this app is hosted from an S3 static website endpoint and not CloudFront or a custom domain, cookie-based auth is a poor fit for the browser flow. The backend already returns a JWT token in the login response, so the clean fix was to use that token from the frontend.

## Root cause

### Sign up

The register page already sends the correct payload to `/auth/register`, so the form itself was not the issue.

The real issue was backend CORS. The backend allowed a single hardcoded origin like:

```text
http://dev-starttech-frontend-ee4128bc.s3-website-us-east-1.amazonaws.com
```

but Terraform creates the frontend bucket with a random suffix:

```tf
bucket = "${var.environment}-starttech-frontend-${random_id.suffix.hex}"
```

That means the current S3 website URL can change, and when it no longer matches the hardcoded origin, the browser blocks the request.

### Sign in

The login endpoint returns both:

1. A `user`
2. A `token`

But the frontend was relying on `httpOnly` cookies for follow-up requests like `/users/me`, `/tasks`, and profile updates.

With static S3 website hosting, the frontend and backend are cross-origin. That makes cookie-based browser auth fragile here, especially compared with using the JWT the backend already returns.

## What I changed

### 1. Backend CORS fix

File changed:

- `starttech-application/backend/internal/middleware/cors.go`
- `starttech-infra/terraform/modules/compute/userdata.sh`

What I changed:

1. Replaced the strict origin-only matching with a custom origin check.
2. Kept explicitly configured allowed origins working.
3. Added support for dynamic StartTech S3 website origins that match the static hosting pattern.
4. Removed the stale hardcoded S3 website origin from `userdata.sh` and left only `http://localhost:5173` for local development.

Why:

This fixes sign up and sign in requests coming from the current S3 website URL even when the bucket suffix changes, and it prevents the deployment config from advertising an outdated frontend origin.

### 2. Frontend token-based auth fix

File changed:

- `starttech-application/frontend/src/lib/apiClient.ts`

What I changed:

1. Added token storage helpers using `localStorage`.
2. Added an Axios request interceptor.
3. Automatically send `Authorization: Bearer <token>` when a token exists.

Why:

This removes the frontend’s dependency on cross-site cookies for authenticated API calls and makes sign in work reliably with static hosting.

### 3. Login flow update

File changed:

- `starttech-application/frontend/src/routes/login.tsx`

What I changed:

1. Stored the JWT token returned by `/auth/login`.
2. Kept the existing user state update and redirect behavior.

Why:

The backend already returns a usable token. Saving it on successful login lets the rest of the frontend authenticate future requests correctly.

### 4. Auth bootstrap and logout update

File changed:

- `starttech-application/frontend/src/context/AuthContext.tsx`

What I changed:

1. On app load, only call `/users/me` when a saved token exists.
2. Clear the token if fetching the current user fails.
3. Clear the token on logout success.
4. Also clear the token on logout error so the client never gets stuck in a half-signed-in state.

Why:

This keeps the frontend auth state consistent and makes sign in, refresh, and sign out behave correctly.

## What I did

I traced the full frontend auth path and found that:

1. The sign-up form payload was already correct.
2. The backend route was already correct.
3. The CORS configuration was too rigid for dynamic S3 static website URLs.
4. The frontend sign-in flow was depending on cookies when it should have been using the JWT already returned by the backend.

I then fixed the backend CORS handling and switched the frontend auth flow to bearer-token requests.
I also removed the old hardcoded S3 website URL from `starttech-infra/terraform/modules/compute/userdata.sh` so the deployment no longer carries a stale origin value.

## Why this fixes both flows

### Sign up now works because

The browser request is no longer blocked just because the current S3 website bucket URL changed.

### Sign in now works because

After login, the frontend stores the returned token and sends it on all authenticated API calls, so requests like `/users/me` and `/tasks` do not depend on cross-site cookies.

## Verification

### Verified

I ran:

```bash
npx tsc --noEmit -p tsconfig.app.json
```

and it completed successfully.

### Could not fully verify in this environment

1. `vite build` was blocked by a local `esbuild` spawn permission issue in the sandbox.
2. Go test commands were blocked by local permission issues around Go telemetry/cache paths in this environment.

So the frontend TypeScript layer was verified, but full runtime build/test verification was limited by environment restrictions rather than by the code changes themselves.

## CloudWatch check after deployment

CloudWatch log group:

```text
/starttech/backend
```

After deploying the backend change:

1. Open CloudWatch Logs.
2. Open `/starttech/backend`.
3. Attempt sign up and sign in from the S3 website frontend.
4. Confirm requests now reach the backend normally.

If there is still any backend-side issue after this fix, CloudWatch should now show the real server response instead of the browser failing early at the CORS/auth boundary.

# フロントエンド認証連携

Email/PasswordとGoogleの認証画面とOAuthフローはFirebase Client SDKで処理します。
ログイン後にFirebase ID tokenを取得し、バックエンドへBearer tokenとして送信してください。

## Firebase Console

AuthenticationのSign-in methodで次のプロバイダーを有効化します。

- Email/Password
- Google

## Web SDK

```ts
import { initializeApp } from "firebase/app";
import {
  createUserWithEmailAndPassword,
  getAuth,
  GoogleAuthProvider,
  signInWithEmailAndPassword,
  signInWithPopup,
  type User,
} from "firebase/auth";

const firebaseApp = initializeApp({
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
  appId: import.meta.env.VITE_FIREBASE_APP_ID,
});

export const auth = getAuth(firebaseApp);

export async function signUpWithEmail(email: string, password: string) {
  const credential = await createUserWithEmailAndPassword(auth, email, password);
  return connectToBackend(credential.user);
}

export async function signInWithEmail(email: string, password: string) {
  const credential = await signInWithEmailAndPassword(auth, email, password);
  return connectToBackend(credential.user);
}

export async function signInWithGoogle() {
  const credential = await signInWithPopup(auth, new GoogleAuthProvider());
  return connectToBackend(credential.user);
}

async function connectToBackend(user: User) {
  const idToken = await user.getIdToken();
  const response = await fetch(
    `${import.meta.env.VITE_API_BASE_URL}/api/v1/auth/login`,
    {
      method: "POST",
      headers: { Authorization: `Bearer ${idToken}` },
    },
  );

  if (!response.ok) {
    throw new Error(`Backend authentication failed: ${response.status}`);
  }
  return response.json();
}
```

以降の認証必須APIでも、同じ形式で最新のID tokenをAuthorizationヘッダーへ設定します。

```ts
const idToken = await auth.currentUser?.getIdToken();
const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/me`, {
  headers: { Authorization: `Bearer ${idToken}` },
});
```

本番では `VITE_API_BASE_URL` をHTTPSのAPI URLへ設定し、そのフロントエンドOriginだけを
バックエンドの `CORS_ALLOWED_ORIGINS` に指定してください。

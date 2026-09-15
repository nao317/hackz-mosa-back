# 認証アーキテクチャ方針

## 採用方式

認証はFirebase Authentication、アプリケーションユーザーの永続化はPostgreSQLで行う。
Go APIは独自アクセストークンやCookieセッションを発行せず、Firebase ID tokenをBearer tokenとして
検証するステートレス方式を採用する。

```text
Browser
  -> Firebase Authentication (Email / Google)
  <- Firebase ID token
  -> Go API (Authorization: Bearer <ID token>)
       -> Firebase Admin SDK (署名・期限・失効・無効化を検証)
       -> PostgreSQL (Firebase UIDをキーにユーザーをupsert)
```

## 各レイヤーの責務

| レイヤー | 責務 |
| --- | --- |
| フロントエンド | 認証UI、Email/Password入力、Google OAuth、ID token更新、Firebase sign-out |
| Firebase Authentication | credential検証、アカウント連携、ID token発行、パスワードリセット |
| Go API | ID token検証、許可provider判定、API認可、アプリユーザー同期 |
| PostgreSQL | アプリ内ユーザーID、プロフィール、利用中provider、将来追加するアプリ固有データ |

パスワード、Googleのaccess token、Firebase ID token、サービスアカウントJSONはPostgreSQLへ
保存しない。サービスアカウントJSONは実行環境からApplication Default Credentialsとして注入する。

## API契約

### `POST /api/v1/auth/login`

Firebaseへのログイン直後と、ブラウザで認証状態を復元した直後に呼ぶ。ID tokenを検証し、
PostgreSQLのユーザーを作成または最新プロフィールへ更新する。何度呼んでも同じFirebase UIDに対して
同じアプリ内ユーザーを返す。

### `GET /api/v1/me`

Bearer tokenから現在のアプリ内ユーザーを返す。将来の認証必須APIも同じID token検証境界を通す。

成功・エラーを問わず認証レスポンスは `Cache-Control: no-store` とし、ブラウザや中間キャッシュに
ユーザー情報を保存させない。

## フロントエンドとの整合

- React Router開発サーバーは `http://localhost:5173`、Go APIは `http://localhost:8080`。
- `VITE_API_BASE_URL=http://localhost:8080` をローカル開発の既定値とする。
- Go APIの `CORS_ALLOWED_ORIGINS` は5173とDocker実行時の3000を許可する。
- Firebaseの `AuthUser` は認証UI表示用、Go APIが返すUserはアプリ固有データの基準とする。
- UIはGo APIでのtoken検証とユーザー同期が成功してから認証済み状態へ遷移する。
- ログアウトはFirebase Client SDKの `signOut` で行う。Go APIはサーバーセッションを持たない。

## セキュリティ方針

- 本番通信はHTTPSのみとする。
- 許可Originは本番フロントエンドの完全なOriginに限定し、ワイルドカードを使わない。
- Firebase Admin SDKで失効済みtokenと無効ユーザーも検査する。
- EmailとGoogle以外の `sign_in_provider` はGo APIで拒否する。
- 認証エラーの内部詳細やFirebase SDKのエラー本文はクライアントへ返さない。
- Firebase UID以外のクライアント申告ユーザーIDは認可判断に利用しない。

## 今後の拡張判断

- 認証必須APIが増えた時点で、ID token検証とユーザー解決をEcho middlewareへ共通化する。
- ロールや管理者権限はFirebase custom claimsを検証し、サーバー側use caseでも認可する。
- 全端末ログアウトが必要になった場合だけ、refresh token失効用APIを追加する。
- Firebase Auth EmulatorをCIへ追加し、ブラウザからPostgreSQLまでのE2Eテストを構築する。
- プロフィール編集を追加する場合、Firebase管理項目とアプリ固有項目の所有境界を先に定義する。

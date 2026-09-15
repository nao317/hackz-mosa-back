# hackz-mosa-back

Go、Echo v5、Firebase Authentication、PostgreSQLで動くAPIサーバーです。

FirebaseがEmail/Password、Google、Appleの認証フローを担当し、このAPIはFirebase ID tokenを
検証してアプリケーションユーザーをPostgreSQLへ同期します。パスワードやOAuth credentialは
バックエンドに保存しません。

## Requirements

- Go 1.26+
- Docker
- Firebaseプロジェクト
- Firebase Admin SDKのサービスアカウントJSON

## Firebase setup

Firebase ConsoleのAuthenticationでEmail/Password、Google、Appleを有効化してください。
サービスアカウントJSONは `secrets/firebase-service-account.json` に配置します。このディレクトリは
Gitの管理対象外です。

Apple側で必要なService ID、Team ID、Key ID、秘密鍵、Return URLの設定を含むフロントエンド連携は
[docs/frontend-integration.md](docs/frontend-integration.md) を参照してください。

## Start

`.env.example` を元に `.env` を作り、Firebase project IDを設定します。

```sh
docker compose up --build
```

Goを直接起動する場合はPostgreSQLを先に起動します。

```sh
make db-up
set -a
. ./.env
set +a
make run
```

## API

### `POST /api/v1/auth/login`

Firebaseログイン後のID tokenを検証し、ユーザーをPostgreSQLへ作成または更新します。

```http
Authorization: Bearer <Firebase ID token>
```

### `GET /api/v1/me`

同じBearer tokenを検証し、現在のユーザーを返します。どちらのエンドポイントもEmail、Google、Apple
以外のFirebase providerは拒否します。

### `GET /health`

コンテナのliveness確認用エンドポイントです。

## Configuration

| Environment variable | Required | Description |
| --- | --- | --- |
| `DATABASE_URL` | Yes | PostgreSQL接続文字列 |
| `FIREBASE_PROJECT_ID` | Yes | Firebase project ID |
| `GOOGLE_APPLICATION_CREDENTIALS` | Yes | サービスアカウントJSONのパス |
| `CORS_ALLOWED_ORIGINS` | No | 許可Originのカンマ区切り。既定値は `http://localhost:3000` |
| `PORT` | No | HTTPポート。既定値は `8080` |

`FIREBASE_AUTH_EMULATOR_HOST` を設定するとFirebase Auth Emulatorも利用できます。

## Architecture

```text
cmd/api                  Composition root
internal/domain          Entity and boundary interfaces
internal/usecase         Authentication application logic
internal/adapter/firebase  Firebase token verification
internal/adapter/postgres  User persistence and schema
internal/adapter/http      Echo handlers
internal/server          Routing and middleware
```

## Check

```sh
make check
```

フォーマット、静的解析、race detector付きテスト、ビルドを実行します。GitHub Actionsでは
PostgreSQLを使うrepository統合テスト、Go moduleファイル、Dockerイメージも検証します。

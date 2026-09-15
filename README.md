# hackz-mosa-back

Go と Echo v5 で動く API サーバーです。

## Requirements

- Go 1.26+
- Docker（コンテナで起動する場合）

## Start

ローカルで起動します。

```sh
make run
```

Docker Compose を使う場合は次のコマンドで起動します。

```sh
docker compose up --build
```

起動後は以下のエンドポイントを利用できます。

- `GET http://localhost:8080/`
- `GET http://localhost:8080/health`

ポートを変更する場合は `PORT` 環境変数を指定してください。

```sh
PORT=3000 make run
```

## Check

```sh
make check
```

`make check` はフォーマット、静的解析、race detector 付きテスト、ビルドを実行します。
GitHub Actions ではこれらに加え、Go module ファイルと Docker イメージのビルドも検証します。

# マッピングAPI連携

すべてのエンドポイントでFirebase ID tokenが必要です。登録は認証ユーザーごとに分離されます。

```http
Authorization: Bearer <Firebase ID token>
Content-Type: application/json
```

## エリアと曲を登録する

```http
POST /api/v1/map-mappings

{
  "area": [
    { "latitude": 35.681, "longitude": 139.766 },
    { "latitude": 35.682, "longitude": 139.768 },
    { "latitude": 35.680, "longitude": 139.769 }
  ],
  "track": {
    "id": "audius-track-id",
    "title": "Track title",
    "artist": "Artist name",
    "artworkUrl": "https://...",
    "audiusUrl": "https://audius.co/...",
    "durationSeconds": 180,
    "streamUrl": "https://api.audius.co/v1/tracks/audius-track-id/stream",
    "source": "audius"
  }
}
```

`area`は3点以上500点以下で指定します。最後に始点を重ねて送信した場合はサーバーが重複点を除去します。
成功時は`201 Created`で`{ "mapping": ... }`を返します。

## 登録済みエリアを取得する

```http
GET /api/v1/map-mappings
```

作成日時の新しい順に`{ "mappings": [...] }`を返します。

## 登録を削除する

```http
DELETE /api/v1/map-mappings/{mappingId}
```

成功時は`204 No Content`です。他ユーザーのIDまたは存在しないIDは`404 not_found`を返します。

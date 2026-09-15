# プレイリストAPI連携

すべてのエンドポイントでFirebase ID tokenが必要です。

```http
Authorization: Bearer <Firebase ID token>
Content-Type: application/json
```

APIのベースパスは `/api/v1` です。プレイリストは認証ユーザーごとに分離され、
別ユーザーのプレイリストIDを指定しても `404 not_found` を返します。

## プレイリストを作成する

```http
POST /api/v1/playlists

{
  "name": "朝のプレイリスト",
  "description": "通勤中に聴く曲"
}
```

`name` は必須で100文字まで、`description` は任意で500文字までです。
成功時は `201 Created` で `{ "playlist": ... }` を返します。

## 一覧と詳細を取得する

```http
GET /api/v1/playlists
GET /api/v1/playlists/{playlistId}
```

一覧は `{ "playlists": [...] }`、詳細は `{ "playlist": ... }` です。
どちらも `tracks` を再生順で含みます。一覧は最終更新が新しい順です。

## Audiusの曲を追加する

フロントエンドの `PlayableTrack` をそのままbodyへ渡せます。

```http
POST /api/v1/playlists/{playlistId}/tracks

{
  "id": "audius-track-id",
  "title": "Track title",
  "artist": "Artist name",
  "artworkUrl": "https://...",
  "audiusUrl": "https://audius.co/...",
  "durationSeconds": 180,
  "streamUrl": "https://api.audius.co/v1/tracks/audius-track-id/stream",
  "source": "audius"
}
```

成功時は `201 Created` で更新後のプレイリストを返します。同じAudius曲の重複追加は
`409 duplicate_track`、31曲目は `422 playlist_full` です。

レスポンス内の曲には `id` と `trackId` があります。`id` はプレイリスト内の項目IDで、
並べ替えと削除に使います。`trackId` はリクエストの `PlayableTrack.id` に対応します。

```json
{
  "id": 81,
  "trackId": "audius-track-id",
  "position": 1
}
```

再生時に `PlaylistTrack` を `PlayableTrack` として扱う場合は、`trackId` を `id` に戻します。

## 曲を並べ替える

現在の全曲について、プレイリスト項目の `id` を希望順に送ります。追加・削除と競合して
古い一覧を送った場合は `422 invalid_track_order` になるため、最新のプレイリストを再取得してください。

```http
PUT /api/v1/playlists/{playlistId}/tracks/order

{
  "itemIds": [83, 81, 82]
}
```

成功時は `200 OK` で更新後のプレイリストを返します。

## 曲を削除する

```http
DELETE /api/v1/playlists/{playlistId}/tracks/{itemId}
```

成功時は `200 OK` で更新後のプレイリストを返します。

## プレイリストを削除する

```http
DELETE /api/v1/playlists/{playlistId}
```

成功時は `204 No Content` です。所属する曲も同時に削除されます。

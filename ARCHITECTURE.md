# URL短縮サービス アーキテクチャドキュメント

## 概要

このURL短縮サービスは、Domain Driven Design (DDD) の原則に基づいて設計された、スケーラブルで拡張可能なアプリケーションです。長いURLを短い識別子にマッピングし、高速なリダイレクト機能を提供します。

## アーキテクチャ全体図

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Presentation   │    │   Application   │    │     Domain      │
│     Layer       │───▶│     Layer       │───▶│     Layer       │
│                 │    │                 │    │                 │
│ HTTP Handlers   │    │   Services      │    │   Entities      │
│ REST API        │    │   Use Cases     │    │   Repositories  │
└─────────────────┘    └─────────────────┘    │   Interfaces    │
                                              └─────────────────┘
                                                       │
                                              ┌─────────────────┐
                                              │ Infrastructure  │
                                              │     Layer       │
                                              │                 │
                                              │ Memory Repo     │
                                              │ Base62 KGS      │
                                              │ Analytics       │
                                              └─────────────────┘
```

## レイヤー構成

### 1. Domain Layer (ドメイン層)
**パッケージ:** `internal/shorturl/domain/`

ビジネスロジックの中核を担当し、外部の技術的な実装に依存しない純粋なビジネスルールを定義します。

#### エンティティ
- **ShortURL**: 短縮URLのドメインエンティティ
  - ID生成とバリデーション
  - 有効期限チェック
  - アクティブ状態管理

#### インターフェース
- **ShortURLRepository**: データ永続化の抽象化
- **KeyGenerationService**: 一意ID生成の抽象化  
- **AnalyticsService**: 分析データ送信の抽象化

#### バリューオブジェクト
- **AnalyticsEvent**: 分析イベントの不変オブジェクト

### 2. Application Layer (アプリケーション層)
**パッケージ:** `internal/shorturl/app/`

ユースケースを調整し、ドメインオブジェクト間の相互作用を管理します。

#### サービス
- **ShortURLService**: 主要なアプリケーションサービス
  - 短縮URL作成
  - 長いURL取得
  - 分析イベント送信

#### DTOs (Data Transfer Objects)
- **CreateShortURLRequest/Response**: URL作成用
- **GetLongURLRequest**: URL取得用
- **ShortURLResponse**: レスポンス用

### 3. Infrastructure Layer (インフラストラクチャ層)
**パッケージ:** `internal/shorturl/infra/`

技術的な実装の詳細を担当し、外部システムとの統合を行います。

#### 実装
- **MemoryShortURLRepository**: インメモリデータストア
- **Base62KeyGenerationService**: Base62エンコーディングによるID生成
- **MockAnalyticsService**: 分析イベント送信のモック実装

### 4. Presentation Layer (プレゼンテーション層)
**パッケージ:** `internal/shorturl/interfaces/http/`

HTTP APIエンドポイントを提供し、外部からのリクエストを処理します。

#### ハンドラー
- **ShortURLHandler**: RESTful API エンドポイント
  - `POST /v1/createShortUrl`
  - `GET /v1/getLongUrl`
  - `GET /<shortId>` (リダイレクト)

## 主要コンポーネント

### Key Generation Service (KGS)

READMEの設計に基づいて実装された、一意ID生成サービスです。

```go
type KeyGenerationService interface {
    GenerateUniqueID() (string, error)
    GetMultipleIDs(count int) ([]string, error)
}
```

**特徴:**
- Base62エンコーディング (0-9, a-z, A-Z)
- 7文字長の短縮ID生成
- カウンター基盤でのスケーラブルな実装
- バッファリング機能による高速レスポンス

### Analytics Pipeline

分析データの収集と送信を担当するコンポーネントです。

**イベントタイプ:**
- `url_created`: 短縮URL作成時
- `url_accessed`: 短縮URLアクセス時
- `url_deactivated`: 短縮URL非活性化時

### Repository Pattern

データアクセスを抽象化し、ドメイン層を技術的な実装から分離します。

```go
type ShortURLRepository interface {
    Save(shortURL *ShortURL) error
    FindByID(id string) (*ShortURL, error)
    FindAll() ([]*ShortURL, error)
    Delete(id string) error
}
```

## API仕様

### 短縮URL作成
```http
POST /v1/createShortUrl
Content-Type: application/json

{
    "longUrl": "https://example.com/very/long/path",
    "customUrl": "custom123",
    "expiry": "2024-12-31T23:59:59Z",
    "userMetadata": {
        "userId": "user123",
        "campaign": "winter2024"
    }
}
```

### 長いURL取得
```http
GET /v1/getLongUrl
Content-Type: application/json

{
    "shortUrl": "http://localhost:8080/abc1234",
    "userMetadata": {
        "source": "api"
    }
}
```

### リダイレクト
```http
GET /<shortId>
→ 302 Found
Location: https://example.com/very/long/path
```

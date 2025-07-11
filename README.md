# Short URL Service

高性能でスケーラブルなURL短縮サービス

## 📋 目次

- [概要](#概要)
- [クイックスタート](#クイックスタート)
- [開発環境セットアップ](#開発環境セットアップ)
- [プロジェクト構成](#プロジェクト構成)
- [API仕様](#api仕様)
- [開発ワークフロー](#開発ワークフロー)
- [テスト](#テスト)
- [デプロイ](#デプロイ)
- [トラブルシューティング](#トラブルシューティング)

## 概要

このプロジェクトは、Domain Driven Design (DDD) の原則に基づいて構築されたURL短縮サービスです。
長いURLを短い識別子にマッピングし、高速なリダイレクト機能を提供します。

### 主要機能

- ✅ 長いURLの短縮URL生成
- ✅ 短縮URLからのリダイレクト
- ✅ カスタムURL指定対応
- ✅ 有効期限設定
- ✅ 分析データ収集
- ✅ 管理者機能（一覧表示、無効化）

### 技術スタック

- **言語**: Go 1.24.3
- **アーキテクチャ**: Domain Driven Design (DDD)
- **Web**: 標準net/httpライブラリ
- **データストア**: インメモリ（本番では外部DB想定）
- **ID生成**: Base62エンコーディング
- **CI/CD**: GitHub Actions
- **静的解析**: golangci-lint

## クイックスタート

### 前提条件

- Go 1.24.3 以上
- Git
- Make（推奨）

### 1. プロジェクトのクローン

```bash
git clone https://github.com/oharai/short-url.git
cd short-url
```

### 2. 依存関係の解決

```bash
make deps
```

### 3. アプリケーションの起動

```bash
make run
```

サーバーが起動し、`http://localhost:8080` でアクセス可能になります。

### 4. 動作確認

```bash
# 短縮URL作成
curl -X POST http://localhost:8080/v1/createShortUrl \
  -H "Content-Type: application/json" \
  -d '{"longUrl": "https://example.com"}'

# レスポンス例:
# {"shortUrl": "http://localhost:8080/abc1234"}

# リダイレクト確認
curl -L http://localhost:8080/abc1234
```

## 開発環境セットアップ

### 必要なツールのインストール

```bash
# 開発ツールの一括インストール
make install-tools

# 個別インストール（必要に応じて）
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### 開発用設定ファイル

プロジェクトルートに以下の設定ファイルが含まれています：

- `.golangci.yml` - 静的解析設定
- `Makefile` - 開発コマンド集
- `Dockerfile` - コンテナ化設定
- `.github/workflows/` - CI/CD設定

## プロジェクト構成

```
short-url/
├── cmd/api/                 # アプリケーションエントリーポイント
│   └── main.go             # HTTPサーバー起動とDI設定
├── internal/shorturl/       # メインビジネスロジック
│   ├── domain/             # ドメイン層（ビジネスルール）
│   │   ├── shorturl.go     # ShortURLエンティティ
│   │   ├── repository.go   # リポジトリインターフェース
│   │   ├── kgs.go         # Key Generation Service
│   │   └── analytics.go    # 分析サービス
│   ├── app/                # アプリケーション層（ユースケース）
│   │   ├── service.go      # ビジネスロジックの調整
│   │   └── dto.go         # データ転送オブジェクト
│   ├── infra/              # インフラストラクチャ層（技術詳細）
│   │   ├── memory_repository.go     # インメモリ実装
│   │   ├── base62_kgs.go           # Base62 ID生成
│   │   └── mock_analytics.go       # モック分析サービス
│   └── interfaces/http/    # プレゼンテーション層（API）
│       └── handler.go      # HTTPハンドラー
├── test/                   # 統合テスト
│   └── integration_test.go
├── docs/                   # ドキュメント
├── .github/workflows/      # GitHub Actions
├── Dockerfile             # コンテナ設定
├── Makefile              # 開発コマンド
├── .golangci.yml         # 静的解析設定
└── README.md             # このファイル
```

### アーキテクチャ詳細

詳細なアーキテクチャ情報は [ARCHITECTURE.md](./ARCHITECTURE.md) を参照してください。

## API仕様

### 短縮URL作成

```http
POST /v1/createShortUrl
Content-Type: application/json

{
    "longUrl": "https://example.com/very/long/path",
    "customUrl": "custom123",           // オプション
    "expiry": "2024-12-31T23:59:59Z",  // オプション
    "userMetadata": {                   // オプション
        "userId": "user123",
        "campaign": "winter2024"
    }
}
```

**レスポンス**:
```json
{
    "shortUrl": "http://localhost:8080/abc1234"
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

**レスポンス**: 302 Found + Location ヘッダー

### リダイレクト

```http
GET /<shortId>
```

**レスポンス**: 302 Found + Location ヘッダー

### 管理者API

```http
# 全URL一覧取得
GET /admin/shorturls

# URL無効化
DELETE /admin/deactivate?id=<shortId>
```

## 開発ワークフロー

### 日常的な開発コマンド

```bash
# ヘルプ表示
make help

# コード整形
make fmt

# 静的解析
make lint

# テスト実行
make test

# カバレッジ確認
make coverage

# 全品質チェック
make check

# ビルド
make build

# 統合テスト
make integration-test

# CI/CD相当のローカル実行
make ci
```

### コーディング規約

1. **gofmt**: コードフォーマットを統一
2. **golangci-lint**: 静的解析による品質保証
3. **テストカバレッジ**: 90%以上を維持
4. **ドキュメント**: 公開関数にはGoDocコメント必須

### プルリクエストフロー

1. **ブランチ作成**: `feature/your-feature-name`
2. **開発**: 機能実装とテスト作成
3. **品質チェック**: `make pre-commit`実行
4. **プルリクエスト作成**: GitHub上でPR作成
5. **CI/CD**: 自動テストとビルド
6. **コードレビュー**: チームレビュー後マージ

### Git コミット規約

```
feat: 新機能追加
fix: バグ修正
docs: ドキュメント更新
style: フォーマット修正
refactor: リファクタリング
test: テスト追加/修正
chore: その他の変更
```

## テスト

### テスト種別

1. **ユニットテスト**: 個別コンポーネントのテスト
2. **統合テスト**: コンポーネント間の連携テスト
3. **ベンチマークテスト**: パフォーマンステスト

### テスト実行

```bash
# 全テスト
make test

# カバレッジ付きテスト
make test-coverage

# 統合テスト
make integration-test

# ベンチマークテスト
make benchmark

# HTMLカバレッジレポート
make coverage-html
```

### テストファイル配置

- ユニットテスト: 各パッケージ内に `*_test.go`
- 統合テスト: `test/` ディレクトリ
- モック: 各層の `infra/` 内に配置

### カバレッジ目標

- **全体**: 90%以上
- **ドメイン層**: 95%以上
- **アプリケーション層**: 90%以上
- **インフラ層**: 85%以上

## デプロイ

### Docker使用

```bash
# Dockerイメージ作成
make docker-build

# Dockerコンテナ起動
make docker-run
```

### 環境変数

本番環境では以下の環境変数を設定：

```bash
# アプリケーション設定
PORT=8080
BASE_URL=https://yourdomain.com

# データベース設定（将来実装予定）
DB_HOST=localhost
DB_PORT=5432
DB_NAME=shorturl
DB_USER=username
DB_PASSWORD=password

# 分析設定（将来実装予定）
ANALYTICS_ENDPOINT=https://analytics.yourdomain.com
ANALYTICS_API_KEY=your-api-key
```

### GitHub Actions

Push/PR時に以下が自動実行されます：

- ✅ 静的解析 (golangci-lint)
- ✅ セキュリティスキャン (gosec)
- ✅ テスト実行
- ✅ カバレッジチェック
- ✅ ビルド確認
- ✅ 依存関係脆弱性チェック

## トラブルシューティング

### よくある問題

#### 1. ビルドエラー

```bash
# 依存関係の更新
make deps
go mod tidy

# キャッシュクリア
go clean -modcache
```

#### 2. テスト失敗

```bash
# 詳細なテスト実行
go test -v -race ./...

# 特定パッケージのテスト
go test -v ./internal/shorturl/domain/
```

#### 3. 静的解析エラー

```bash
# golangci-lint詳細実行
golangci-lint run --verbose

# 自動修正可能な項目の修正
golangci-lint run --fix
```

#### 4. カバレッジ不足

```bash
# カバレッジ詳細表示
make coverage

# HTMLレポートでカバレッジ箇所確認
make coverage-html
open coverage.html
```

### デバッグ設定

開発時のログレベル調整やデバッグ情報出力については、`cmd/api/main.go` を参照してください。

### パフォーマンス調査

```bash
# プロファイリング有効化でのベンチマーク
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof

# プロファイル分析
go tool pprof cpu.prof
go tool pprof mem.prof
```

## コントリビューション

1. Issue作成またはコメント
2. ブランチ作成: `git checkout -b feature/your-feature`
3. 実装: コード作成、テスト追加
4. 品質チェック: `make pre-commit`
5. プルリクエスト作成

詳細は [CONTRIBUTING.md](./CONTRIBUTING.md) を参照（将来作成予定）。

## ライセンス

このプロジェクトは [LICENSE](./LICENSE) ファイルに記載されたライセンスの下で提供されています。

## サポート

質問や問題がある場合：

1. [GitHub Issues](https://github.com/oharai/short-url/issues) で報告
2. プロジェクト内ドキュメントを確認
3. チーム内での相談

---

**開発チームへの参加を歓迎します！** 🚀
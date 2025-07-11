# golangci-lint セットアップガイド

## golangci-lintとは

golangci-lintは、Go言語用の高速で設定可能なリンターです。複数のリンターを統合し、並列実行により高速な静的解析を提供します。

## インストール

### 1. バイナリのインストール

#### macOS (Homebrew)
```bash
brew install golangci-lint
```

#### Linux/macOS (curl)
```bash
# 最新版をインストール
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 特定のバージョンをインストール
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
```

#### Go install
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

#### Windows (PowerShell)
```powershell
# 最新版をインストール
iwr -useb https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | iex
```

### 2. Docker経由での実行
```bash
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run
```

## 基本的な使用方法

### 1. プロジェクト全体の解析
```bash
golangci-lint run
```

### 2. 特定のディレクトリの解析
```bash
golangci-lint run ./internal/...
```

### 3. 特定のファイルの解析
```bash
golangci-lint run file.go
```

### 4. 利用可能なリンターの確認
```bash
golangci-lint linters
```

### 5. 設定の確認
```bash
golangci-lint config
```

## 設定ファイルの解説

プロジェクトの `.golangci.yml` は初心者向けの設定になっています：

### 有効になっているリンター（初期設定）

- **errcheck**: エラーハンドリングの確認
- **gosimple**: コードの簡略化提案
- **govet**: Go公式の検査ツール
- **ineffassign**: 無効な代入の検出
- **staticcheck**: 静的解析
- **typecheck**: 型チェック
- **unused**: 未使用の要素検出
- **gofmt**: フォーマットチェック
- **goimports**: import文の整理
- **misspell**: スペルミスの検出
- **unconvert**: 不要な型変換の検出
- **unparam**: 未使用パラメーターの検出
- **gocyclo**: 循環的複雑度の測定
- **nolintlint**: nolintディレクティブの検証
- **gosec**: セキュリティ問題の検出

### 段階的に有効化可能なリンター

設定ファイルのコメントアウト部分には、より厳格なリンターが記載されています：

```yaml
# コメントアウトされているリンター例
# - bodyclose       # HTTPレスポンスボディのクローズチェック
# - dupl            # コードの重複検出
# - funlen          # 関数の長さチェック
# - goconst         # 定数化可能な文字列の検出
# - lll             # 行の長さチェック
```

## 実行例

### 1. 基本的な実行
```bash
# プロジェクトルートで実行
golangci-lint run

# 出力例
internal/shorturl/app/service.go:45:2: ineffectual assignment to err (ineffassign)
internal/shorturl/domain/shorturl.go:123:1: exported function `NewShortURL` should have comment (missing-doc)
```

### 2. 自動修正（サポートされているリンターのみ）
```bash
golangci-lint run --fix
```

### 3. 特定のリンターのみ実行
```bash
golangci-lint run --enable-only=errcheck,gosec
```

### 4. 特定のリンターを無効化
```bash
golangci-lint run --disable=gosec
```

### 5. 詳細情報付きで実行
```bash
golangci-lint run -v
```

## IDE統合

### VS Code
1. Go拡張機能をインストール
2. 設定に以下を追加：
```json
{
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"]
}
```

### GoLand/IntelliJ
1. Settings → Tools → Go Tools → golangci-lint
2. パスを設定し、有効化

### Vim/Neovim
ALE, vim-go, coc-go などのプラグインでサポートされています。

## CI/CD統合

### GitHub Actions
```yaml
- name: golangci-lint
  uses: golangci/golangci-lint-action@v3
  with:
    version: latest
```

### Make統合
```makefile
lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix
```

## トラブルシューティング

### 1. メモリ不足エラー
```bash
# 並列数を制限
golangci-lint run --concurrency=2

# または環境変数で設定
export GOLANGCI_LINT_CACHE=/tmp/golangci-cache
```

### 2. 特定の問題を無視
```go
//nolint:gosec // セキュリティ警告を無視する理由
func unsafeFunction() {
    // 危険なコード
}
```

### 3. キャッシュクリア
```bash
golangci-lint cache clean
```

## 設定のカスタマイズ

### より厳格な設定に移行する手順

1. **段階1**: 現在の設定で問題を修正
2. **段階2**: コメントアウトされたリンターを1つずつ有効化
3. **段階3**: より厳格な設定値に変更

例：
```yaml
# 段階的に厳格化
linters:
  enable:
    # 現在有効なリンター...
    - bodyclose    # 新たに追加
    - dupl         # 新たに追加

linters-settings:
  gocyclo:
    min-complexity: 10  # 15 → 10 に変更
  lll:
    line-length: 100    # 120 → 100 に変更
```

## ベストプラクティス

1. **段階的導入**: 一度にすべてのリンターを有効にせず、段階的に導入
2. **チーム合意**: リンターの設定はチーム全体で合意を取る
3. **CI統合**: Pull Request時に自動チェックを実行
4. **定期見直し**: プロジェクトの成熟度に応じて設定を見直し
5. **ドキュメント化**: 特定のリンターを無効化する理由を記録

## 参考リンク

- [golangci-lint公式ドキュメント](https://golangci-lint.run/)
- [利用可能なリンター一覧](https://golangci-lint.run/usage/linters/)
- [設定ファイル詳細](https://golangci-lint.run/usage/configuration/)
- [GitHub Actions統合](https://github.com/golangci/golangci-lint-action)
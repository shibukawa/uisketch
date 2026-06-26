# uisketch

`uisketch` は、`.uisketch.md` のレイアウトを描画・編集するための Go + Web ワークスペースです。Web エディタ、ローカルプロジェクト編集、CLI レンダリング、Wails のデスクトップ起動を同じソースから扱います。

## Web エディタ

GitHub Pages でホストされた Web エディタを利用できます。

https://shibukawa.github.io/uisketch/

## セットアップ

```bash
npm install
go test ./...
```

## よく使うコマンド

### Web エディタを開発モードで起動

```bash
npm run dev:web
```

Vite の開発サーバーが起動し、ブラウザで編集できます。ビルド済みの確認だけなら次も使えます。

```bash
npm run build:web
npm run preview:web
```

### CLI で描画する

```bash
go run ./cmd/uisketch render examples/sample.uisketch.md
go run ./cmd/uisketch render --format ascii examples/sample.uisketch.md
```

標準入力も使えます。

```bash
cat examples/sample.uisketch.md | go run ./cmd/uisketch render
```

### Markdown を再生成する

```bash
go run ./cmd/uisketch markdown --output out.md examples/input.md
go run ./cmd/uisketch markdown --overwrite examples/input.md
```

### ローカルプロジェクトを編集する

```bash
go run ./cmd/uisketch edit --project . --no-open
go run ./cmd/uisketch edit --project . path/to/file.uisketch.md
```

`edit` は指定した project root の中だけを読み書きします。ファイル名を渡すと、そのファイルを初期状態で開きます。

### デスクトップアプリを作る

Wails 版は同じ Web UI と Go のファイルサービスを使います。

```bash
wails dev
```

ビルドだけ行う場合は次です。

```bash
wails build
```

`wails` コマンドが PATH にない場合は、同等のビルドを次で実行できます。

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build -nosyncgomod -nopackage
```

Wails ビルドでは `scripts/build-wails-frontend.sh` が呼ばれ、`packages/web/dist` の成果物が `desktop_assets` にコピーされます。
埋め込み用の Web 成果物だけ更新したい場合は `npm run build:desktop-assets` を使えます。

## 追加の確認コマンド

```bash
npm run test
npm run typecheck
npm run build
```

## モードの使い分け

- ブラウザモード: `localStorage` に名前付き保存、読み込み、削除ができます。
- ローカルプロジェクトモード: `uisketch edit` または Wails 版で、プロジェクト内ファイルを直接読み書きします。
- 共有 URL はブラウザモード向けです。ローカルプロジェクトモードでは使いません。

## 関連仕様

- [UI Sketch File Format](specs/ui-sketch-file-format.md)
- [Uisketch CLI](specs/uisketch-cli.md)
- [Web Visual Editor](specs/web-visual-editor.md)

## エージェント向けスキル

このリポジトリは、エージェント向けの Codex スキルを [.agents/skills/uisketch-skill](.agents/skills/uisketch-skill) で提供しています。このスキルは `.uisketch.md` ファイルの理解、作成、検証、修正、実装ブリーフとしての利用を支援します。

## ライセンス

このプロジェクトは GNU Affero General Public License v3.0 or later でライセンスされています。詳細は [LICENSE](LICENSE) を参照してください。

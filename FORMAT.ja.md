# uisketch format

この文書は、このリポジトリにおける `uisketch` レイアウトソースの正規データ構造を説明します。

## 対象範囲

正規の保存形式は、`type: uisketch` の frontmatter を持ち、1 つ以上の明示的な `uisketch` fenced block を含む `.uisketch.md` 文書です。レイアウト木そのものは YAML です。

この形式は次の用途で共通です。

- ブラウザモードの保存
- ローカルプロジェクトファイル
- CLI の描画
- Markdown 埋め込みワークフロー
- Wails デスクトップ編集

## 文書構造

レイアウトの観点では、文書は 1 つの root node と、その下に連なる子ノードの木です。

代表的な root surface は次です。

- `browser`
- `window`
- `mobile`
- `dialog`
- `menu`

例:

```yaml
browser:
  id: account-editor
  title: Account Editor
  address: https://example.local/accounts/42
  children:
    - vstack:
        children:
          - label: Account
          - input:
              id: name
              label: Name
          - hstack:
              children:
                - spacer:
                - button:
                    label: Save
                    action: save-account
```

外側のキーが root surface の型です。その下はすべて semantic component tree です。

## 共通フィールド

パーサーは、存在する場合に次のフィールドを保持します。

| フィールド | 意味 |
| --- | --- |
| `id` | 選択、リンク、ツール連携のための安定した要素 ID。 |
| `title` | `browser`、`window`、`mobile`、`dialog` のようなタイトル付き root surface の表示名。 |
| `label` | コントロールや内容ノードの可視ラベル。 |
| `action` | インタラクティブなコントロールに付く action ID。 |
| `anchor` | 別定義や参照先へのナビゲーション目標。 |
| `address` | browser surface のアドレスまたは URL。 |
| `hint` | input 系コントロール向けの短い補助文。 |
| `children` | 入れ子の semantic child nodes。 |
| `labels` | tab label など、コンポーネント定義上のラベル集合。 |
| `columns` | コンポーネント型に応じた列ラベルまたは grid 列数。 |
| `widths` | 横方向の stack sizing。 |
| `heights` | 縦方向の stack sizing。 |
| `prompt` | ノードに付ける AI 向けの自由記述ガイド。 |
| `data` | ノードに付ける自由形式メタデータ。 |

`prompt` は人間または AI 向けのガイドです。可視 UI としては描画されません。

`data` は構造化または半構造化のメタデータです。これも可視 UI としては描画されません。

どちらのフィールドも parse / save の往復で保持されます。

## ノード種別

このリポジトリでは、ピクセル配置ではなく semantic layout tree を使います。

通常、`children` を持つ container node は次のようなものです。

- `browser`
- `window`
- `mobile`
- `dialog`
- `menu`
- `vstack`
- `hstack`
- `grid`
- `splitter`
- `tabs`
- `section`

通常、visible content や振る舞いを持つ leaf node は次のようなものです。

- `button`
- `toggle`
- `input`
- `textarea`
- `combobox`
- `checkbox`
- `radio`
- `label`
- `table`
- `image`
- `custom`
- `list`
- `spacer`

実際の対応関係は Go の parser、renderer、editor inspector に実装されています。ファイル形式は、モデルが増えても追加の semantic field を保持できるよう、ある程度許容的に作られています。

## Tabs

`tabs` は tab label と active body を別々に持ちます。

active body は `children` に、tab label は `labels` に置きます。

## Prompt と Data

すべての node は `prompt` と `data` を持てます。

- `prompt` は AI 向けのメモや指示を書き込むための textarea 相当の項目です。
- `data` はプロジェクト固有の情報を入れる自由形式メタデータです。

どちらも単独で描画には影響しません。保存時には node と一緒に保持され、読み込み時に復元されます。

## ラウンドトリップ規則

- root surface type を保持する。
- child の順序を保持する。
- `prompt` と `data` を保持する。
- エディタや serializer が扱える範囲では、未知の semantic field も保持する。
- selection、drag state、zoom、cursor position などの transient state は保存しない。

正規の保存結果は、エディタが `.uisketch.md` 文書または project file に書き戻す YAML ソースです。ブラウザモードでは同じソースを `localStorage` に包んで保存でき、ローカルプロジェクトモードでは同じソースをディスク上のファイルとして保存しますが、共通の中核表現は YAML の layout tree です。

## 関連仕様

- [UI Sketch File Format](specs/ui-sketch-file-format.md)
- [Uisketch CLI](specs/uisketch-cli.md)
- [Web Visual Editor](specs/web-visual-editor.md)
- [UI Component Catalog](specs/ui-component-catalog.md)
- [UI Layout DSL](specs/ui-layout-dsl.md)

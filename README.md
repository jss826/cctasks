# cctasks

Claude Code の Task List 機能で作成されたタスクを閲覧・編集できる TUI ツール。

## Features

- プロジェクト一覧表示・選択
- タスク一覧（グループ別折りたたみ表示）
- ステータス / グループ / キーワードフィルタ
- タスク作成・編集・削除
- ステータスのクイック変更
- グループ管理（作成・編集・削除・並び替え・色設定）
- ファイル変更の自動検出・更新（操作時）
- キーボードナビゲーション

## Requirements

- Go 1.21+
- Claude Code v2.1.16+ (タスク機能を使用する場合)

## Installation

```bash
go install github.com/jss826/cctasks@latest
```

または、ソースからビルド:

```bash
git clone https://github.com/jss826/cctasks.git
cd cctasks
go build -o cctasks
```

Windows:
```powershell
winget install GoLang.Go
go build -o cctasks.exe
```

## Usage

```bash
./cctasks
```

## Claude Code Task List のセットアップ

Claude Code v2.1.16 以降で Task List 機能を有効にする方法:

1. プロジェクトの `.claude/settings.local.json` に以下を追加:

```json
{
  "env": {
    "CLAUDE_CODE_TASK_LIST_ID": "your-project-name"
  }
}
```

2. タスクは `~/.claude/tasks/your-project-name/` に保存されます

詳細: https://docs.anthropic.com/en/docs/claude-code/interactive-mode#task-list

## Key Bindings

### Project Selection
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `Enter` | Select project |
| `?` | Toggle help |
| `r` | Refresh |
| `q` | Quit |

### Task List
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `Enter` | View details / Toggle group |
| `n` | New task |
| `e` | Edit task |
| `s` | Quick status change |
| `f` | Cycle status filter |
| `g` | Cycle group filter |
| `G` | Manage groups |
| `/` | Search |
| `p` | Back to projects |
| `r` | Refresh |
| `q` | Quit |

### Task Detail
| Key | Action |
|-----|--------|
| `Esc` | Back to list |
| `e` | Edit |
| `s` | Cycle status |
| `d` | Delete |
| `q` | Quit |

### Task Edit
| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `↑/↓` | Change status/group (when focused) |
| `Ctrl+S` | Save |
| `Esc` | Cancel |

### Group Management
| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `Enter` or `e` | Edit group |
| `n` | New group |
| `d` | Delete group |
| `K/J` | Move up/down |
| `Esc` | Back |

## Data Format

タスクは `~/.claude/tasks/<project>/` に個別ファイルとして保存されます:

```
~/.claude/tasks/<project>/
├── 1.json
├── 2.json
├── 3.json
└── _groups.json
```

各タスクファイル (`{id}.json`):

```json
{
  "id": "1",
  "subject": "Task title",
  "description": "Task description",
  "status": "pending",
  "blocks": [],
  "blockedBy": [],
  "owner": "",
  "metadata": {
    "group": "Backend"
  }
}
```

グループ設定 (`_groups.json`):

```json
{
  "groups": [
    {
      "name": "Backend",
      "order": 1,
      "color": "#8b5cf6"
    }
  ]
}
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

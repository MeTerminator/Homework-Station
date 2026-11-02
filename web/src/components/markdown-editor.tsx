import { useTheme } from "@/app/providers/theme-provider";
import MDEditor from "@uiw/react-md-editor/nohighlight";
import * as commands from "@uiw/react-md-editor/commands-cn";
import rehypeSanitize from "rehype-sanitize";

const editorCommands = [
  commands.bold,
  commands.italic,
  commands.strikethrough,
  commands.divider,
  {
    ...commands.title,
    buttonProps: { "aria-label": "插入标题(ctrl + 1)", title: "插入标题(ctrl + 1)" },
  },
  {
    ...commands.quote,
    buttonProps: { "aria-label": "插入引用(ctrl + q)", title: "插入引用(ctrl + q)" },
  },
  commands.hr,
  commands.divider,
  commands.link,
  commands.image,
  commands.code,
  commands.codeBlock,
  commands.divider,
  commands.unorderedListCommand,
  commands.orderedListCommand,
  commands.checkedListCommand,
  commands.table,
];
const extraCommands = [
  commands.codeEdit,
  commands.codeLive,
  commands.codePreview,
  commands.divider,
  commands.fullscreen,
];
const previewOptions = { rehypePlugins: [rehypeSanitize] };

export function MarkdownEditor({
  value,
  onChange,
  label,
  placeholder,
}: {
  value: string;
  onChange: (value: string) => void;
  label: string;
  placeholder?: string;
}) {
  const { dark } = useTheme();
  return (
    <div className="markdown-editor" data-color-mode={dark ? "dark" : "light"}>
      <MDEditor
        // The built-in overflow handler persists a dialog's computed scroll lock
        // as an inline body style. Let CSS own the fullscreen lock instead.
        overflow={false}
        value={value}
        onChange={(next) => onChange(next ?? "")}
        commands={editorCommands}
        extraCommands={extraCommands}
        preview="live"
        previewOptions={previewOptions}
        height={360}
        minHeight={240}
        autoFocus={false}
        defaultTabEnable
        textareaProps={{ "aria-label": label, placeholder }}
      />
    </div>
  );
}

import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { UploadBox } from "@/features/attachments/components/upload-box";
import { fileURL } from "@/lib/api/urls";
import type { Assignment, Attachment } from "@/types/models";
import { Code, Eye } from "lucide-react";
import { useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { AssignmentFieldChange } from "../types";
export function AssignmentDescriptionEditor({
  form,
  onChangeField,
  files,
  setFiles,
  setUploading,
}: {
  form: Assignment;
  onChangeField: AssignmentFieldChange;
  files: Attachment[];
  setFiles: (files: Attachment[]) => void;
  setUploading: (busy: boolean) => void;
}) {
  const [preview, setPreview] = useState(false);
  return (
    <Panel title="作业描述">
      <div className="editor-tabs">
        <div className="row">
          <Button
            variant={preview ? "ghost" : "secondary"}
            size="sm"
            onClick={() => setPreview(false)}
          >
            <Code size={14} />
            编辑
          </Button>
          <Button
            variant={preview ? "secondary" : "ghost"}
            size="sm"
            onClick={() => setPreview(true)}
          >
            <Eye size={14} />
            预览
          </Button>
        </div>
        <span className="muted small">支持 Markdown / GFM</span>
      </div>
      {preview ? (
        <div className="markdown panel" style={{ minHeight: 230 }}>
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {form.description || "暂无作业描述。"}
          </ReactMarkdown>
        </div>
      ) : (
        <Textarea
          aria-label="作业 Markdown 描述"
          rows={12}
          style={{ fontFamily: "monospace", lineHeight: 1.8 }}
          value={form.description}
          onChange={(e) => onChangeField("description", e.target.value)}
          placeholder={"# 作业说明\n\n写下作业目标、要求以及参考资料…"}
        />
      )}
      <div style={{ marginTop: 18 }}>
        <UploadBox
          files={files}
          onBusy={setUploading}
          onChange={(next) => {
            const added = next.filter((f) => !files.some((x) => x.id === f.id));
            setFiles(next);
            if (added.length)
              onChangeField(
                "description",
                form.description +
                  "\n" +
                  added
                    .map((f) => `[${f.name.replace(/[\[\]\\]/g, "")}](${fileURL(f.id)})`)
                    .join("\n"),
              );
          }}
        />
      </div>
    </Panel>
  );
}

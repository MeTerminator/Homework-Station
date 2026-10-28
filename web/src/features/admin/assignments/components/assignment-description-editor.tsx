import { MarkdownEditor } from "@/components/markdown-editor";
import { Panel } from "@/components/panel";
import { UploadBox } from "@/features/attachments/components/upload-box";
import { fileURL } from "@/lib/api/urls";
import type { Assignment, Attachment } from "@/types/models";
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
  return (
    <Panel title="作业描述">
      <MarkdownEditor
        label="作业 Markdown 描述"
        value={form.description}
        onChange={(value) => onChangeField("description", value)}
        placeholder={"# 作业说明\n\n写下作业目标、要求以及参考资料…"}
      />
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

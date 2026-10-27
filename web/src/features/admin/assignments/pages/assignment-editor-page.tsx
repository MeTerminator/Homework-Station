import { Confirm } from "@/components/feedback/confirm";
import { LoadState } from "@/components/feedback/load-state";
import { PageHeading } from "@/components/page-heading";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { blankAssignment } from "@/features/admin/assignments/lib/defaults";
import { SubmissionHistory } from "@/features/assignments/components/submission-history";
import { useData } from "@/hooks/use-data";
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import type { Assignment, Attachment, Category, Settings } from "@/types/models";
import { Save, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { AssignmentBasicInfo } from "../components/assignment-basic-info";
import { AssignmentDescriptionEditor } from "../components/assignment-description-editor";
import { AssignmentSubmissionRules } from "../components/assignment-submission-rules";
export function AssignmentEditor() {
  const { id } = useParams();
  const create = id === "new";
  const nav = useNavigate();
  const [form, setForm] = useState<Assignment>(blankAssignment);
  const [files, setFiles] = useState<Attachment[]>([]);
  const [busy, setBusy] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [confirm, setConfirm] = useState(false);
  const cats = useData(() => api<Category[]>("/categories"));
  const q = useData(async () => {
    if (create) {
      const s = await api<Settings>("/admin/settings");
      return {
        assignment: {
          ...blankAssignment,
          max_size: +s.default_max_size,
          extensions: s.default_extensions,
        },
        attachments: [] as Attachment[],
      };
    }
    return api<{ assignment: Assignment; attachments: Attachment[] }>(
      `/assignments/${id}`,
    );
  }, [id]);
  useEffect(() => {
    if (q.data) {
      setForm(q.data.assignment);
      setFiles(q.data.attachments);
    }
  }, [q.data]);
  const set = <K extends keyof Assignment>(k: K, v: Assignment[K]) =>
    setForm((f) => ({ ...f, [k]: v }));
  const save = () => {
    setBusy(true);
    perform(async () => {
      const a = await api<Assignment>(
        create ? "/admin/assignments" : `/admin/assignments/${id}`,
        json(create ? "POST" : "PUT", {
          name: form.name,
          category_id: form.category_id,
          status: form.status,
          deadline: form.deadline,
          sort: form.sort,
          description: form.description,
          allow_attachments: form.allow_attachments,
          extensions: form.extensions,
          max_size: form.max_size,
          multiple: form.multiple,
          withdraw: form.withdraw,
          attachment_ids: files.map((f) => f.id),
        }),
      );
      nav(`/admin/assignments/${a.id}`, { replace: true });
      q.reload();
    }, "作业已保存").finally(() => setBusy(false));
  };
  return (
    <>
      <PageHeading
        title={create ? "创建作业" : "编辑作业"}
        description="完善作业内容与提交规则，让收集更轻松。"
      >
        <div className="row">
          <Link to="/admin/assignments">
            <Button variant="outline">返回列表</Button>
          </Link>
          <Button disabled={busy || uploading || q.loading} onClick={save}>
            <Save size={16} />
            {busy ? "保存中…" : "保存作业"}
          </Button>
        </div>
      </PageHeading>
      <LoadState {...q} retry={q.reload}>
        <div className="stack">
          <AssignmentBasicInfo form={form} onChangeField={set} cats={cats.data || []} />
          <AssignmentDescriptionEditor
            form={form}
            onChangeField={set}
            files={files}
            setFiles={setFiles}
            setUploading={setUploading}
          />
          <AssignmentSubmissionRules form={form} onChangeField={set} />
          {!create && (
            <>
              <SubmissionHistory assignmentID={Number(id)} />
              <Panel>
                <div className="row" style={{ justifyContent: "space-between" }}>
                  <span className="muted small">删除前需先清除全部历史提交。</span>
                  <Button variant="destructive" onClick={() => setConfirm(true)}>
                    <Trash2 size={15} />
                    删除作业
                  </Button>
                </div>
              </Panel>
            </>
          )}
        </div>
      </LoadState>
      <Confirm
        open={confirm}
        onClose={() => setConfirm(false)}
        title="删除此作业？"
        description="作业信息将永久删除，描述中的附件仍保留在附件库。"
        onConfirm={async () => {
          await api(`/admin/assignments/${id}`, json("DELETE"));
          nav("/admin/assignments");
        }}
      />
    </>
  );
}

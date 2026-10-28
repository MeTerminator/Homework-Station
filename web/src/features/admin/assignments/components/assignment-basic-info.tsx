import { Field } from "@/components/form/field";
import { Panel } from "@/components/panel";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { CategorySelect } from "@/features/categories/components/category-select";
import { localDateTime } from "@/lib/format";
import type { Assignment, Category } from "@/types/models";
import type { AssignmentFieldChange } from "../types";
export function AssignmentBasicInfo({
  form,
  onChangeField,
  cats,
}: {
  form: Assignment;
  onChangeField: AssignmentFieldChange;
  cats: Category[];
}) {
  return (
    <Panel title="基本信息">
      <div className="form-grid">
        <Field label="作业名称">
          <Input
            value={form.name}
            maxLength={200}
            onChange={(e) => onChangeField("name", e.target.value)}
            placeholder="输入作业名称"
          />
        </Field>
        <Field label="所属分类">
          <CategorySelect
            cats={cats}
            value={form.category_id}
            onChange={(v) => onChangeField("category_id", v)}
          />
        </Field>
        <Field label="作业简介（可选）" hint="显示在作业卡片中，最多 200 字。" full>
          <Textarea
            value={form.summary}
            maxLength={200}
            rows={3}
            onChange={(e) => onChangeField("summary", e.target.value)}
            placeholder="简要介绍本次作业的内容或目标"
          />
        </Field>
        <Field label="发布状态">
          <select
            value={form.status}
            onChange={(e) => onChangeField("status", e.target.value)}
          >
            <option value="hidden">隐藏</option>
            <option value="published">已发布</option>
            <option value="closed">已截止</option>
          </select>
        </Field>
        <Field label="截止时间（可选）" hint="留空则长期可交。时间按浏览器所在时区显示。">
          <Input
            type="datetime-local"
            value={form.deadline ? localDateTime(form.deadline) : ""}
            onChange={(e) =>
              onChangeField(
                "deadline",
                e.target.value ? new Date(e.target.value).toISOString() : null,
              )
            }
          />
        </Field>
        <Field label="排序" hint="数字越大越靠后。">
          <Input
            type="number"
            value={form.sort}
            onChange={(e) => onChangeField("sort", +e.target.value)}
          />
        </Field>
      </div>
    </Panel>
  );
}

import { Status } from "@/features/assignments/components/assignment-status";
import { categoryLabel } from "@/features/categories/lib/category-label";
import { dateText } from "@/lib/format";
import type { Assignment, Category } from "@/types/models";
import { ChevronRight } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
export function AssignmentTable({
  items,
  cats,
}: {
  items: Assignment[];
  cats: Category[];
}) {
  const nav = useNavigate();
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            {[
              "ID",
              "作业名称",
              "作业分类",
              "状态",
              "提交人数",
              "截止日期",
              "创建日期",
              "",
            ].map((s, i) => (
              <th key={i}>{s}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {items.map((a) => (
            <tr
              key={a.id}
              className="click-row"
              onClick={() => nav(`/admin/assignments/${a.id}`)}
            >
              <td className="muted">#{a.id.toString().padStart(3, "0")}</td>
              <td className="table-title">
                <Link to={`/admin/assignments/${a.id}`}>{a.name}</Link>
              </td>
              <td>
                <span className="badge" style={{ whiteSpace: "normal" }}>
                  {categoryLabel(a.category_id, cats)}
                </span>
              </td>
              <td>
                <Status status={a.status} />
              </td>
              <td>
                {a.submitters} <span className="muted">人</span>
              </td>
              <td className="muted">{dateText(a.deadline, true)}</td>
              <td className="muted">{dateText(a.created_at, true)}</td>
              <td>
                <ChevronRight size={15} className="muted" />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

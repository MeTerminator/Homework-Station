import { Status } from "@/features/assignments/components/assignment-status";
import type { AssignmentCategoryGroup } from "@/features/categories/lib/category-tree";
import { dateText } from "@/lib/format";
import { ArrowRight, BookOpen, Clock } from "lucide-react";
import { Link } from "react-router-dom";
export function AssignmentCategorySection({ group }: { group: AssignmentCategoryGroup }) {
  const headingTags = ["h2", "h3", "h4", "h5", "h6"] as const;
  const cardTags = ["h3", "h4", "h5", "h6"] as const;
  const Heading = headingTags[Math.min(group.depth, headingTags.length - 1)];
  const CardHeading = cardTags[Math.min(group.depth, cardTags.length - 1)];
  return (
    <section
      className="assignment-category"
      data-depth={group.depth + 1}
      aria-labelledby={`category-heading-${group.id}`}
    >
      <div className="category-title">
        <Heading aria-level={group.depth + 2} id={`category-heading-${group.id}`}>
          {group.name}
        </Heading>
        <span className="category-count" title="含子分类的作业总数">
          {group.total}
        </span>
      </div>
      {group.items.length > 0 && (
        <div className="assignment-grid">
          {group.items.map((a) => (
            <Link to={`/assignments/${a.id}`} className="assignment-card" key={a.id}>
              <div className="row" style={{ justifyContent: "space-between" }}>
                <span
                  className="stat-icon"
                  style={{ width: 34, height: 34, borderRadius: 8 }}
                >
                  <BookOpen size={18} />
                </span>
                {a.submitted ? (
                  <span className="badge green">已提交</span>
                ) : (
                  <Status status={a.status} />
                )}
              </div>
              <CardHeading aria-level={group.depth + 3}>{a.name}</CardHeading>
              <p className="small muted" style={{ lineHeight: 1.7 }}>
                {a.allow_attachments ? "支持文本与附件提交" : "纯文本提交"} ·{" "}
                {a.multiple ? "允许多次提交" : "单次提交"}
              </p>
              <div className="card-bottom">
                <span className="row" style={{ gap: 5 }}>
                  <Clock size={12} />
                  {dateText(a.deadline, true)}
                </span>
                <ArrowRight size={15} />
              </div>
            </Link>
          ))}
        </div>
      )}
      {group.children.length > 0 && (
        <div className="category-children">
          {group.children.map((child) => (
            <AssignmentCategorySection key={child.id} group={child} />
          ))}
        </div>
      )}
    </section>
  );
}

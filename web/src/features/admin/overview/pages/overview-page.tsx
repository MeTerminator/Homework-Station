import { useSession } from "@/app/providers/session-provider";
import { Empty } from "@/components/feedback/empty";
import { LoadState } from "@/components/feedback/load-state";
import { PageHeading } from "@/components/page-heading";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { AssignmentTable } from "@/features/admin/assignments/components/assignment-table";
import { useData } from "@/hooks/use-data";
import { api } from "@/lib/api/client";
import { sizeText } from "@/lib/format";
import type { Assignment, Category, Page } from "@/types/models";
import {
  ArrowUpRight,
  ClipboardCheck,
  HardDrive,
  Paperclip,
  Plus,
  Users,
} from "lucide-react";
import { Link } from "react-router-dom";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
export interface OverviewData {
  users: number;
  attachments: number;
  capacity: number;
  submissions: number;
  days: {
    date: string;
    users: number;
    attachments: number;
    capacity: number;
    submissions: number;
  }[];
  capacity_note: string;
}

export function Overview() {
  const q = useData(() => api<OverviewData>("/admin/overview"));
  const latest = useData(() => api<Page<Assignment>>("/admin/assignments?size=10"));
  const cats = useData(() => api<Category[]>("/categories"));
  const { user } = useSession();
  return (
    <>
      <PageHeading
        title="总览"
        description={`欢迎回来，${user?.username}。这里是你的作业收集概况。`}
      >
        <Link to="/admin/assignments/new">
          <Button>
            <Plus size={16} />
            创建作业
          </Button>
        </Link>
      </PageHeading>
      <LoadState {...q} retry={q.reload}>
        {q.data && (
          <>
            <div className="stats">
              {[
                {
                  label: "用户总数",
                  value: q.data.users,
                  Icon: Users,
                  hint: "已注册的全部用户",
                },
                {
                  label: "附件总数",
                  value: q.data.attachments,
                  Icon: Paperclip,
                  hint: "包含上传中的附件",
                },
                {
                  label: "附件容量",
                  value: sizeText(q.data.capacity),
                  Icon: HardDrive,
                  hint: "SHA256 去重后的实际容量",
                },
                {
                  label: "作业提交份数",
                  value: q.data.submissions,
                  Icon: ClipboardCheck,
                  hint: "包含全部历史提交记录",
                },
              ].map(({ label, value, Icon, hint }) => (
                <div className="stat" key={label}>
                  <div className="stat-top">
                    <span className="muted small">{label}</span>
                    <span className="stat-icon">
                      <Icon size={20} />
                    </span>
                  </div>
                  <div className="stat-value">{value}</div>
                  <div className="stat-caption">{hint}</div>
                </div>
              ))}
            </div>
            <div className="chart-grid">
              {[
                {
                  key: "users",
                  label: "新增用户",
                  color: "#FF406E",
                  unit: "人",
                },
                {
                  key: "submissions",
                  label: "作业提交",
                  color: "#56bfa7",
                  unit: "份",
                },
                {
                  key: "attachments",
                  label: "新增附件",
                  color: "#e7ad66",
                  unit: "个",
                },
                {
                  key: "capacity",
                  label: "新增附件容量",
                  color: "#7cace3",
                  unit: "MB",
                },
              ].map(({ key, label, color, unit }) => (
                <Panel
                  key={key}
                  className="chart-panel"
                  title={label}
                  action={<span className="badge gray">近 30 天 · UTC</span>}
                >
                  <div className="chart-box">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart
                        data={q.data!.days.map((d) => ({
                          ...d,
                          capacity: +(d.capacity / 1024 ** 2).toFixed(2),
                        }))}
                        margin={{ top: 22, right: 8, bottom: 0, left: -23 }}
                      >
                        <defs>
                          <linearGradient id={"fill-" + key} x1="0" y1="0" x2="0" y2="1">
                            <stop offset="0%" stopColor={color} stopOpacity={0.2} />
                            <stop offset="100%" stopColor={color} stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid
                          vertical={false}
                          stroke="var(--border)"
                          strokeDasharray="4 4"
                        />
                        <XAxis
                          dataKey="date"
                          tickFormatter={(s) => s.slice(5)}
                          tick={{
                            fontSize: 10,
                            fill: "var(--muted-foreground)",
                          }}
                          axisLine={false}
                          tickLine={false}
                          minTickGap={35}
                        />
                        <YAxis
                          tick={{
                            fontSize: 10,
                            fill: "var(--muted-foreground)",
                          }}
                          allowDecimals={key === "capacity"}
                          axisLine={false}
                          tickLine={false}
                        />
                        <Tooltip
                          contentStyle={{
                            background: "var(--card)",
                            border: "1px solid var(--border)",
                            borderRadius: 8,
                            fontSize: 12,
                          }}
                          formatter={(v) => [`${v} ${unit}`, label]}
                        />
                        <Area
                          type="monotone"
                          dataKey={key}
                          stroke={color}
                          strokeWidth={2.5}
                          fill={`url(#fill-${key})`}
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </Panel>
              ))}
            </div>
            <p className="muted small" style={{ margin: "-10px 0 25px" }}>
              {q.data.capacity_note}
            </p>
          </>
        )}
      </LoadState>
      <Panel
        title="近期作业"
        action={
          <Link to="/admin/assignments" className="link small row">
            查看全部
            <ArrowUpRight size={14} />
          </Link>
        }
      >
        <LoadState {...latest} retry={latest.reload}>
          {latest.data?.items.length ? (
            <AssignmentTable
              items={latest.data.items.slice(0, 5)}
              cats={cats.data || []}
            />
          ) : (
            <Empty
              title="还没有发布作业"
              description="创建第一份作业，开始收集同学们的学习成果。"
            >
              <Link to="/admin/assignments/new">
                <Button variant="outline" size="sm">
                  <Plus size={15} />
                  创建作业
                </Button>
              </Link>
            </Empty>
          )}
        </LoadState>
      </Panel>
    </>
  );
}

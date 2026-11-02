import { PageHeading } from "@/components/page-heading";
import { SubmissionHistory } from "@/features/assignments/components/submission-history";

export function SubmissionManager() {
  return (
    <>
      <PageHeading
        title="提交管理"
        description="查看所有作业的提交记录，按提交时间从新到旧排列，包含已撤回的提交。点击附件可预览和下载。"
      />
      <SubmissionHistory />
    </>
  );
}

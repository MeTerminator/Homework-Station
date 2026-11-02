export function Status({ status }: { status: string }) {
  return (
    <span
      className={`badge ${status === "published" ? "green" : status === "closed" ? "red" : "gray"}`}
    >
      {status === "published" ? "已发布" : status === "closed" ? "已截止" : "隐藏"}
    </span>
  );
}

export function ReviewStatus({
  status,
  withdrawn = false,
}: {
  status?: string;
  withdrawn?: boolean;
}) {
  return (
    <span
      className={`badge ${withdrawn ? "gray" : status === "pending" ? "red" : "green"}`}
    >
      {withdrawn
        ? "已撤回"
        : status === "reviewed"
          ? "已评价"
          : status === "pending"
            ? "待回复"
            : "已提交"}
    </span>
  );
}

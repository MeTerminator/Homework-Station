export function Status({ status }: { status: string }) {
  return (
    <span
      className={`badge ${status === "published" ? "green" : status === "closed" ? "red" : "gray"}`}
    >
      {status === "published" ? "已发布" : status === "closed" ? "已截止" : "隐藏"}
    </span>
  );
}

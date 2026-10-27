import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-react";
export function Pager({
  page,
  size,
  total,
  onPage,
  onSize,
}: {
  page: number;
  size: number;
  total: number;
  onPage: (n: number) => void;
  onSize: (n: number) => void;
}) {
  return (
    <div className="pagination">
      <div className="row">
        共 {total} 条
        <select
          aria-label="每页条数"
          value={size}
          onChange={(e) => {
            onSize(+e.target.value);
            onPage(1);
          }}
        >
          {[10, 20, 50, 100].map((n) => (
            <option key={n} value={n}>
              {n} 条 / 页
            </option>
          ))}
        </select>
      </div>
      <div className="row">
        <Button
          variant="outline"
          size="icon"
          disabled={page <= 1}
          onClick={() => onPage(page - 1)}
          aria-label="上一页"
        >
          <ChevronLeft size={16} />
        </Button>
        <span>
          {page} / {Math.max(1, Math.ceil(total / size))}
        </span>
        <Button
          variant="outline"
          size="icon"
          disabled={page * size >= total}
          onClick={() => onPage(page + 1)}
          aria-label="下一页"
        >
          <ChevronRight size={16} />
        </Button>
      </div>
    </div>
  );
}

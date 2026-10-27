import { useTheme } from "@/app/providers/theme-provider";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Check, Monitor, Moon, Sun } from "lucide-react";
export function ThemeSwitch() {
  const { theme, setTheme } = useTheme();
  const Icon = theme === "system" ? Monitor : theme === "light" ? Sun : Moon;
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="切换主题">
          <Icon size={18} />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {(["system", "light", "dark"] as const).map((t, i) => (
          <DropdownMenuItem key={t} onClick={() => setTheme(t)}>
            {[<Monitor size={15} />, <Sun size={15} />, <Moon size={15} />][i]}
            {["跟随系统", "浅色模式", "深色模式"][i]}
            {theme === t && <Check size={14} />}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

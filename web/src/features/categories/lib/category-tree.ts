import type { Assignment, Category } from "@/types/models";

export interface CategoryNode {
  category: Category;
  depth: number;
  children: CategoryNode[];
}

export function categoryTree(categories: Category[]): CategoryNode[] {
  const sorted = [...categories].sort((a, b) => a.sort - b.sort || a.id - b.id);
  const ids = new Set(sorted.map((c) => c.id));
  const visited = new Set<number>();
  const walk = (category: Category, depth: number): CategoryNode => {
    visited.add(category.id);
    return {
      category,
      depth,
      children: sorted
        .filter((c) => c.parent_id === category.id && !visited.has(c.id))
        .map((c) => walk(c, depth + 1)),
    };
  };
  const roots = sorted
    .filter((c) => c.parent_id === null || !ids.has(c.parent_id))
    .map((c) => walk(c, 0));
  // Preserve visibility of legacy orphaned/cyclic records without recursing forever.
  for (const category of sorted)
    if (!visited.has(category.id)) roots.push(walk(category, 0));
  return roots;
}
export function flattenCategories(nodes: CategoryNode[]): CategoryNode[] {
  return nodes.flatMap((node) => [node, ...flattenCategories(node.children)]);
}
export function descendantIDs(nodes: CategoryNode[]): number[] {
  return flattenCategories(nodes).map((node) => node.category.id);
}
export function matchingCategoryIDs(
  nodes: CategoryNode[],
  selected: number[],
): Set<number> {
  const selectedIDs = new Set(selected);
  const matching = new Set<number>();
  const visit = (node: CategoryNode, ancestors: number[], inherited: boolean) => {
    const isSelected = selectedIDs.has(node.category.id);
    if (isSelected) ancestors.forEach((id) => matching.add(id));
    if (isSelected || inherited) matching.add(node.category.id);
    for (const child of node.children)
      visit(child, [...ancestors, node.category.id], isSelected || inherited);
  };
  nodes.forEach((node) => visit(node, [], false));
  return matching;
}
export function parentCategoryOptions(
  categories: Category[],
  movingID?: number,
): CategoryNode[] {
  const tree = categoryTree(categories);
  const nodes = flattenCategories(tree);
  const moving = nodes.find((n) => n.category.id === movingID);
  const excluded = new Set(moving ? descendantIDs([moving]) : []);
  return nodes.filter((n) => !excluded.has(n.category.id));
}
export interface AssignmentCategoryGroup {
  id: number;
  name: string;
  depth: number;
  items: Assignment[];
  children: AssignmentCategoryGroup[];
  total: number;
}
export function groupAssignments(
  categories: Category[],
  assignments: Assignment[],
): AssignmentCategoryGroup[] {
  const sortedItems = (id: number | null) =>
    assignments
      .filter((a) => a.category_id === id)
      .sort(
        (a, b) =>
          Number(a.submitted) - Number(b.submitted) || a.sort - b.sort || b.id - a.id,
      );
  const visit = (node: CategoryNode): AssignmentCategoryGroup => {
    const children = node.children.map(visit).filter((g) => g.total > 0);
    const items = sortedItems(node.category.id);
    return {
      id: node.category.id,
      name: node.category.name,
      depth: node.depth,
      items,
      children,
      total: items.length + children.reduce((n, child) => n + child.total, 0),
    };
  };
  const groups = categoryTree(categories)
    .map(visit)
    .filter((g) => g.total > 0);
  const uncategorized = sortedItems(null);
  if (uncategorized.length)
    groups.push({
      id: 0,
      name: "未分类",
      depth: 0,
      items: uncategorized,
      children: [],
      total: uncategorized.length,
    });
  return groups;
}

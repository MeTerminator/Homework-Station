import test from "node:test";
import assert from "node:assert/strict";
import {
  categoryTree,
  flattenCategories,
  parentCategoryOptions,
  groupAssignments,
} from "../src/features/categories/lib/category-tree.ts";

const categories = [
  { id: 1, name: "运维安全", parent_id: null, sort: 0 },
  { id: 2, name: "安全", parent_id: 1, sort: 0 },
  { id: 3, name: "web", parent_id: 2, sort: 0 },
  { id: 4, name: "re", parent_id: 2, sort: 1 },
  { id: 5, name: "运维", parent_id: 1, sort: 1 },
  { id: 6, name: "其他", parent_id: null, sort: 1 },
];
const assignment = (id, category_id, submitted = false, sort = 0) => ({
  id,
  category_id,
  submitted,
  sort,
});

test("categories use tree order and three indentation levels", () => {
  assert.deepEqual(
    flattenCategories(categoryTree([...categories].reverse())).map((n) => [
      n.category.name,
      n.depth,
    ]),
    [
      ["运维安全", 0],
      ["安全", 1],
      ["web", 2],
      ["re", 2],
      ["运维", 1],
      ["其他", 0],
    ],
  );
});
test("new categories can choose parents at any depth", () => {
  assert.deepEqual(
    parentCategoryOptions(categories, 0).map((n) => n.category.id),
    [1, 2, 3, 4, 5, 6],
  );
});
test("moving categories excludes only the category and its descendants", () => {
  assert.deepEqual(
    parentCategoryOptions(categories, 1).map((n) => n.category.id),
    [6],
  );
  assert.deepEqual(
    parentCategoryOptions(categories, 2).map((n) => n.category.id),
    [1, 5, 6],
  );
  assert.deepEqual(
    parentCategoryOptions(categories, 3).map((n) => n.category.id),
    [1, 2, 4, 5, 6],
  );
});
test("deep categories retain indentation, parent options and all ancestor headings", () => {
  const deep = Array.from({ length: 12 }, (_, i) => ({
    id: i + 1,
    name: `分类${i + 1}`,
    parent_id: i === 0 ? null : i,
    sort: 0,
  }));
  assert.equal(flattenCategories(categoryTree(deep)).at(-1).depth, 11);
  assert.equal(parentCategoryOptions(deep, 0).length, 12);
  let group = groupAssignments(deep, [assignment(20, 12)])[0];
  for (let depth = 0; depth < 11; depth++) {
    assert.equal(group.depth, depth);
    group = group.children[0];
  }
  assert.equal(group.items[0].id, 20);
});
test("assignment headings preserve ancestors without direct assignments", () => {
  const groups = groupAssignments(categories, [assignment(11, 3)]);
  assert.equal(groups.length, 1);
  assert.equal(groups[0].name, "运维安全");
  assert.equal(groups[0].items.length, 0);
  assert.equal(groups[0].children[0].name, "安全");
  assert.equal(groups[0].children[0].children[0].name, "web");
  assert.equal(groups[0].children[0].children[0].items[0].id, 11);
  assert.equal(groups[0].total, 1);
});
test("direct assignments are not duplicated; pending submissions retain priority", () => {
  const groups = groupAssignments(categories, [
    assignment(10, 1, true),
    assignment(11, 3),
    assignment(12, 1, false, 50),
    assignment(13, null),
  ]);
  assert.deepEqual(
    groups[0].items.map((a) => a.id),
    [12, 10],
  );
  assert.equal(groups[0].total, 3);
  assert.equal(groups[0].children[0].children[0].total, 1);
  assert.equal(groups[1].name, "未分类");
  assert.equal(groups[1].items[0].id, 13);
});
test("empty category branches disappear under search or pending filters", () => {
  assert.deepEqual(groupAssignments(categories, []), []);
  assert.deepEqual(
    groupAssignments(categories, [assignment(15, 5)]).map((g) =>
      g.children.map((c) => c.name),
    ),
    [["运维"]],
  );
});
test("legacy orphaned and cyclic records remain finite and visible", () => {
  const malformed = [
    { id: 1, name: "A", parent_id: 2, sort: 0 },
    { id: 2, name: "B", parent_id: 1, sort: 0 },
    { id: 3, name: "C", parent_id: 99, sort: 0 },
  ];
  assert.equal(flattenCategories(categoryTree(malformed)).length, 3);
});

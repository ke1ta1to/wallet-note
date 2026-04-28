import { categoriesQuery } from "@/features/category/queries";
import {
  Alert,
  Badge,
  Button,
  Card,
  ColorSwatch,
  Group,
  Skeleton,
  Stack,
  Text,
  Title,
} from "@mantine/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link, Outlet, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs/$orgId/categories")({
  loader: ({ context: { queryClient }, params }) =>
    queryClient.ensureQueryData(categoriesQuery(params.orgId)),
  component: CategoriesLayout,
  pendingComponent: CategoriesPending,
  errorComponent: CategoriesError,
});

function CategoriesLayout() {
  const { orgId } = Route.useParams();
  const { data } = useSuspenseQuery(categoriesQuery(orgId));

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>カテゴリ</Title>
        <Link
          to="/orgs/$orgId/categories/new"
          params={{ orgId }}
          style={{ textDecoration: "none" }}
        >
          <Button component="span">+ 新規作成</Button>
        </Link>
      </Group>
      {data.items.length === 0 ? (
        <Text c="dimmed">まだカテゴリがありません</Text>
      ) : (
        <Stack gap="xs">
          {data.items.map((c) => (
            <Card key={c.id} withBorder padding="sm">
              <Group wrap="nowrap">
                <ColorSwatch color={c.color} size={20} />
                <Text style={{ flex: 1 }}>{c.name}</Text>
                <Badge
                  variant="light"
                  color={c.kind === "income" ? "teal" : "red"}
                >
                  {c.kind === "income" ? "収入" : "支出"}
                </Badge>
              </Group>
            </Card>
          ))}
        </Stack>
      )}
      <Outlet />
    </Stack>
  );
}

function CategoriesPending() {
  return (
    <Stack>
      <Title order={2}>カテゴリ</Title>
      <Skeleton height={48} />
      <Skeleton height={48} />
    </Stack>
  );
}

function CategoriesError({ error }: { error: unknown }) {
  const message =
    (error as { message?: string })?.message ?? "読み込みに失敗しました";
  return (
    <Stack>
      <Title order={2}>カテゴリ</Title>
      <Alert color="red">{message}</Alert>
    </Stack>
  );
}

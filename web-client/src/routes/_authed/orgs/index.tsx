import { meOrgsQuery } from "@/features/organization/queries";
import {
  Alert,
  Badge,
  Button,
  Card,
  Container,
  Group,
  Skeleton,
  Stack,
  Text,
  Title,
} from "@mantine/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link, createFileRoute } from "@tanstack/react-router";
import { format } from "date-fns";

export const Route = createFileRoute("/_authed/orgs/")({
  loader: ({ context: { queryClient } }) =>
    queryClient.ensureQueryData(meOrgsQuery),
  component: Orgs,
  pendingComponent: OrgsPending,
  errorComponent: OrgsError,
});

function Orgs() {
  const { data } = useSuspenseQuery(meOrgsQuery);

  return (
    <Container size="sm" py="xl">
      <Group justify="space-between" mb="md">
        <Title order={1}>Organizations</Title>
        <Button component={Link} to="/orgs/new">
          + 新規作成
        </Button>
      </Group>
      {data.items.length === 0 ? (
        <Text c="dimmed">まだ組織がありません</Text>
      ) : (
        <Stack>
          {data.items.map((m) => (
            <Link
              key={m.org_id}
              to="/orgs/$orgId"
              params={{ orgId: m.org_id }}
              style={{ textDecoration: "none", color: "inherit" }}
            >
              <Card withBorder padding="md">
                <Group justify="space-between">
                  <Text fw={500}>{m.org_name ?? m.org_id}</Text>
                  <Badge variant={m.role === "owner" ? "filled" : "light"}>
                    {m.role}
                  </Badge>
                </Group>
                <Text size="xs" c="dimmed" mt="xs">
                  {format(new Date(m.joined_at), "yyyy年M月d日")} に参加
                </Text>
              </Card>
            </Link>
          ))}
        </Stack>
      )}
    </Container>
  );
}

function OrgsPending() {
  return (
    <Container size="sm" py="xl">
      <Group justify="space-between" mb="md">
        <Title order={1}>Organizations</Title>
        <Button component={Link} to="/orgs/new">
          + 新規作成
        </Button>
      </Group>
      <Stack>
        <Skeleton height={80} />
        <Skeleton height={80} />
      </Stack>
    </Container>
  );
}

function OrgsError({ error }: { error: unknown }) {
  const message =
    (error as { message?: string })?.message ?? "読み込みに失敗しました";
  return (
    <Container size="sm" py="xl">
      <Group justify="space-between" mb="md">
        <Title order={1}>Organizations</Title>
        <Button component={Link} to="/orgs/new">
          + 新規作成
        </Button>
      </Group>
      <Alert color="red">{message}</Alert>
    </Container>
  );
}

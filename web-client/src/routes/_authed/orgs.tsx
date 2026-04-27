import { meOrgsQuery } from "@/features/organization/queries";
import {
  Alert,
  Badge,
  Card,
  Container,
  Group,
  Skeleton,
  Stack,
  Text,
  Title,
} from "@mantine/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { format } from "date-fns";

export const Route = createFileRoute("/_authed/orgs")({
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
      <Title order={1} mb="md">
        Organizations
      </Title>
      {data.items.length === 0 ? (
        <Text c="dimmed">まだ組織がありません</Text>
      ) : (
        <Stack>
          {data.items.map((m) => (
            <Card key={m.org_id} withBorder padding="md">
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
          ))}
        </Stack>
      )}
    </Container>
  );
}

function OrgsPending() {
  return (
    <Container size="sm" py="xl">
      <Title order={1} mb="md">
        Organizations
      </Title>
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
      <Title order={1} mb="md">
        Organizations
      </Title>
      <Alert color="red">{message}</Alert>
    </Container>
  );
}

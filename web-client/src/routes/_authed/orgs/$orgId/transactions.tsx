import { categoriesQuery } from "@/features/category/queries";
import { transactionsQuery } from "@/features/transaction/queries";
import {
  ActionIcon,
  Alert,
  Card,
  ColorSwatch,
  Group,
  Skeleton,
  Stack,
  Text,
  Title,
} from "@mantine/core";
import {
  IconChevronLeft,
  IconChevronRight,
  IconPlus,
} from "@tabler/icons-react";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link, Outlet, createFileRoute } from "@tanstack/react-router";
import { addMonths, format, parse } from "date-fns";
import * as v from "valibot";

const todayMonth = () => format(new Date(), "yyyy-MM");

const SearchSchema = v.object({
  month: v.optional(
    v.fallback(v.pipe(v.string(), v.regex(/^\d{4}-\d{2}$/)), todayMonth),
    todayMonth,
  ),
});

export const Route = createFileRoute("/_authed/orgs/$orgId/transactions")({
  validateSearch: SearchSchema,
  loaderDeps: ({ search: { month } }) => ({ month }),
  loader: ({ context: { queryClient }, params, deps }) =>
    Promise.all([
      queryClient.ensureQueryData(
        transactionsQuery(params.orgId, deps.month),
      ),
      queryClient.ensureQueryData(categoriesQuery(params.orgId)),
    ]),
  component: TransactionsPage,
  pendingComponent: TransactionsPending,
  errorComponent: TransactionsError,
});

function TransactionsPage() {
  const { orgId } = Route.useParams();
  const { month } = Route.useSearch();
  const { data: txs } = useSuspenseQuery(transactionsQuery(orgId, month));
  const { data: cats } = useSuspenseQuery(categoriesQuery(orgId));

  const catById = new Map(cats.items.map((c) => [c.id, c]));
  // Backend returns date ascending; show newest first.
  const items = [...txs.items].reverse();

  const monthDate = parse(month + "-01", "yyyy-MM-dd", new Date());
  const prevMonth = format(addMonths(monthDate, -1), "yyyy-MM");
  const nextMonth = format(addMonths(monthDate, 1), "yyyy-MM");
  const monthLabel = format(monthDate, "yyyy年 M月");

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>取引</Title>
        <Link
          to="/orgs/$orgId/transactions/new"
          params={{ orgId }}
          search={{ month }}
          style={{ textDecoration: "none" }}
          aria-label="取引を追加"
        >
          <ActionIcon size="lg" variant="filled">
            <IconPlus size={18} />
          </ActionIcon>
        </Link>
      </Group>

      <Group justify="space-between">
        <Link
          to="/orgs/$orgId/transactions"
          params={{ orgId }}
          search={{ month: prevMonth }}
          aria-label="前の月"
          style={{ textDecoration: "none" }}
        >
          <ActionIcon component="span" variant="subtle">
            <IconChevronLeft size={18} />
          </ActionIcon>
        </Link>
        <Text fw={500}>{monthLabel}</Text>
        <Link
          to="/orgs/$orgId/transactions"
          params={{ orgId }}
          search={{ month: nextMonth }}
          aria-label="次の月"
          style={{ textDecoration: "none" }}
        >
          <ActionIcon component="span" variant="subtle">
            <IconChevronRight size={18} />
          </ActionIcon>
        </Link>
      </Group>

      {items.length === 0 ? (
        <Text c="dimmed">この月の取引はありません</Text>
      ) : (
        <Stack gap="xs">
          {items.map((t) => {
            const cat = catById.get(t.category_id);
            const isIncome = cat?.kind === "income";
            return (
              <Card key={t.id} withBorder padding="sm">
                <Group wrap="nowrap" align="flex-start">
                  <Stack gap={0} style={{ minWidth: 56 }}>
                    <Text size="sm" c="dimmed">
                      {format(parse(t.date, "yyyy-MM-dd", new Date()), "M/d")}
                    </Text>
                  </Stack>
                  <Stack gap={2} style={{ flex: 1, minWidth: 0 }}>
                    <Group gap={6} wrap="nowrap">
                      {cat && <ColorSwatch color={cat.color} size={12} />}
                      <Text truncate>{cat?.name ?? "(削除済カテゴリ)"}</Text>
                    </Group>
                    {t.memo && (
                      <Text size="xs" c="dimmed" lineClamp={1}>
                        {t.memo}
                      </Text>
                    )}
                  </Stack>
                  <Text fw={600} c={isIncome ? "teal" : "red"}>
                    {isIncome ? "+" : "−"}
                    {t.amount.toLocaleString()}
                  </Text>
                </Group>
              </Card>
            );
          })}
        </Stack>
      )}

      <Outlet />
    </Stack>
  );
}

function TransactionsPending() {
  return (
    <Stack>
      <Title order={2}>取引</Title>
      <Skeleton height={48} />
      <Skeleton height={48} />
      <Skeleton height={48} />
    </Stack>
  );
}

function TransactionsError({ error }: { error: unknown }) {
  const message =
    (error as { message?: string })?.message ?? "読み込みに失敗しました";
  return (
    <Stack>
      <Title order={2}>取引</Title>
      <Alert color="red">{message}</Alert>
    </Stack>
  );
}

import { beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";

import AffiliatesView from "../AffiliatesView.vue";

const {
  listInviters,
  listInviterInvitees,
  showError,
} = vi.hoisted(() => ({
  listInviters: vi.fn(),
  listInviterInvitees: vi.fn(),
  showError: vi.fn(),
}));

vi.mock("@/api/admin/affiliates", () => {
  const api = {
    listInviters,
    listInviterInvitees,
  };
  return {
    affiliatesAPI: api,
    default: api,
  };
});

vi.mock("@/stores", () => ({
  useAppStore: () => ({
    showError,
  }),
}));

vi.mock("@/utils/apiError", () => ({
  extractApiErrorMessage: () => "error",
}));

vi.mock("@/utils/format", () => ({
  formatCurrency: (value: number, currency?: string) => `${currency ?? "USD"}:${value.toFixed(2)}`,
  formatDateTime: (value: string) => `FMT:${value}`,
}));

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  const messages: Record<string, string> = {
    "admin.affiliates.title": "邀请关系",
    "admin.affiliates.description": "查看所有邀请人的邀请关系与返利明细",
    "admin.affiliates.searchPlaceholder": "搜索邀请人邮箱或用户名",
    "admin.affiliates.empty": "暂无邀请关系数据",
    "admin.affiliates.viewInvitees": "查看邀请用户",
    "admin.affiliates.totalLabel": "共 {total} 条",
    "admin.affiliates.inviteesTitle": "{email} 邀请的用户",
    "admin.affiliates.inviteesDescription": "展示该邀请人当前关联的被邀请用户明细。",
    "admin.affiliates.inviteesEmpty": "该邀请人暂无可展示的邀请用户",
    "admin.affiliates.col.email": "邀请人邮箱",
    "admin.affiliates.col.username": "邀请人用户名",
    "admin.affiliates.col.code": "邀请码",
    "admin.affiliates.col.invitedCount": "邀请人数",
    "admin.affiliates.col.totalRebate": "累计返利",
    "admin.affiliates.col.actions": "操作",
    "admin.affiliates.inviteesCol.email": "被邀请用户邮箱",
    "admin.affiliates.inviteesCol.username": "被邀请用户名",
    "admin.affiliates.inviteesCol.joinedAt": "加入时间",
    "admin.affiliates.inviteesCol.totalRebate": "累计返利",
    "pagination.previous": "上一页",
    "pagination.next": "下一页",
    "common.loading": "加载中",
    "common.close": "关闭",
    "common.error": "错误",
  };
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) =>
        (messages[key] ?? key).replace(/\{(\w+)\}/g, (_, token) => params?.[token] ?? `{${token}}`),
    }),
  };
});

describe("AffiliatesView", () => {
  beforeEach(() => {
    listInviters.mockReset();
    listInviterInvitees.mockReset();
    showError.mockReset();

    listInviters.mockResolvedValue({
      items: [
        {
          user_id: 11,
          email: "owner@example.com",
          username: "owner",
          aff_code: "AFFOWNER",
          aff_count: 2,
          total_rebate: 18.8,
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    });
    listInviterInvitees.mockResolvedValue([
      {
        user_id: 101,
        email: "friend@example.com",
        username: "friend",
        created_at: "2026-04-01T00:00:00Z",
        total_rebate: 6.6,
      },
    ]);
  });

  it("loads inviters on mount and shows invitee details on demand", async () => {
    const wrapper = mount(AffiliatesView, {
      global: {
        stubs: {
          AppLayout: { template: "<div><slot /></div>" },
        },
      },
    });

    await flushPromises();

    expect(listInviters).toHaveBeenCalled();
    expect(wrapper.text()).toContain("邀请关系");
    expect(wrapper.text()).toContain("owner@example.com");
    expect(wrapper.text()).toContain("AFFOWNER");
    expect(wrapper.text()).toContain("CNY:18.80");

    const button = wrapper.findAll("button").find((node) => node.text().includes("查看邀请用户"));
    expect(button).toBeDefined();
    await button?.trigger("click");
    await flushPromises();

    expect(listInviterInvitees).toHaveBeenCalledWith(11);
    expect(wrapper.text()).toContain("owner@example.com 邀请的用户");
    expect(wrapper.text()).toContain("friend@example.com");
    expect(wrapper.text()).toContain("FMT:2026-04-01T00:00:00Z");
    expect(wrapper.text()).toContain("CNY:6.60");
  });
});

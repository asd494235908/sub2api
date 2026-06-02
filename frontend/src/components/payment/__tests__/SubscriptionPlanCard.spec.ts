import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { createI18n } from "vue-i18n";
import SubscriptionPlanCard from "../SubscriptionPlanCard.vue";

const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackWarn: false,
  missingWarn: false,
  messages: {
    en: {
      payment: {
        days: "days",
        models: "Models",
        planCard: {
          availableToday: "Remaining today: {count}",
          dailySaleCountdownToEnd: "Ends in {time}",
          dailySaleCountdownToStart: "Starts in {time}",
          dailySaleTime: "Daily sale",
          models: "Models",
          soldOutToday: "Sold out today",
          unavailableNow: "Not available yet",
          quota: "Quota",
          subscriptionTotalLimit: "Per purchase cycle quota",
          cycleTotalShort: "Per purchase quota",
          availableTodayLabel: "Remaining today:",
          availableTodayUnit: "",
          rate: "Rate",
          weeklySaleDays: "Sale days",
          weeklySaleOffDay: "Not on sale today",
          weeklySaleUnlimited: "Every day",
          unlimited: "Unlimited",
        },
        saleUnavailable: "Not available",
        purchaseOnceUnavailable: "Available after current subscription expires",
        subscribeNow: "Subscribe now",
        weekdays: {
          mon: "Mon",
          tue: "Tue",
          wed: "Wed",
          thu: "Thu",
          fri: "Fri",
          sat: "Sat",
          sun: "Sun",
        },
      },
    },
  },
});

const mountPlanCard = (groupPlatform: string) =>
  mount(SubscriptionPlanCard, {
    props: {
      plan: {
        id: 1,
        group_id: 10,
        group_platform: groupPlatform,
        name: "Pro",
        price: 10,
        amount: 1000,
        features: [],
        rate_multiplier: 1,
        validity_days: 30,
        validity_unit: "day",
        supported_model_scopes: ["claude", "gemini_text", "gemini_image"],
        is_active: true,
        for_sale: true,
        daily_purchase_limit: 0,
      },
    },
    global: { plugins: [i18n] },
  });

describe("SubscriptionPlanCard", () => {
  it("does not show Antigravity model scopes for OpenAI plans", () => {
    const text = mountPlanCard("openai").text();

    expect(text).not.toContain("Claude");
    expect(text).not.toContain("Gemini");
    expect(text).not.toContain("Imagen");
  });

  it("shows model scopes for Antigravity plans", () => {
    const text = mountPlanCard("antigravity").text();

    expect(text).toContain("Claude");
    expect(text).toContain("Gemini");
    expect(text).toContain("Imagen");
  });

  it("shows hot supported models first and reveals all models on click", async () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 3,
          group_id: 10,
          group_platform: "openai",
          name: "Pro",
          description: "",
          price: 10,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          for_sale: true,
          daily_purchase_limit: 0,
          supported_models: ["gpt-5.3-codex", "gpt-image-2", "gpt-5.4", "gpt-5.5"],
          sort_order: 1,
        },
      },
      global: {
        plugins: [i18n],
        stubs: {
          Teleport: true,
        },
      },
    });

    const text = wrapper.text();
    expect(text).toContain("payment.planCard.models");
    expect(text).toContain("gpt-5.5");
    expect(text).toContain("gpt-5.4");
    expect(text).toContain("gpt-image-2");
    expect(text).not.toContain("gpt-5.3-codex");
    expect(text).toContain("+1");

    await wrapper.get('[data-test="supported-models-summary"]').trigger("click");

    expect(wrapper.text()).toContain("payment.planCard.supportedModelsTitle");
    expect(wrapper.text()).toContain("gpt-5.3-codex");
  });

  it("does not show supported models row when the plan has no supported models", () => {
    const text = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 4,
          group_id: 10,
          group_platform: "openai",
          name: "Basic",
          description: "",
          price: 10,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          for_sale: true,
          daily_purchase_limit: 0,
          supported_models: [],
          sort_order: 1,
        },
      },
      global: { plugins: [i18n] },
    }).text();

    expect(text).not.toContain("payment.planCard.models");
  });

  it("hides remaining count when daily sale plan is unavailable", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 2,
          group_id: 10,
          group_platform: "openai",
          name: "Flash Sale",
          description: "",
          price: 10,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          for_sale: true,
          daily_purchase_limit: 3,
          daily_purchase_remaining: 1,
          daily_sale_starts_at: "09:00",
          daily_sale_ends_at: "18:00",
          daily_sale_status: "pending",
          daily_sale_countdown_seconds: 3661,
          daily_sale_available_for_payment: false,
          sort_order: 1,
        },
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("payment.planCard.dailySaleTime");
    expect(wrapper.text()).toContain("09:00 - 18:00");
    expect(wrapper.text()).toContain("payment.planCard.dailySaleCountdownToStart");
    expect(wrapper.text()).not.toContain("payment.planCard.availableTodayLabel");
    const button = wrapper.get("button");
    expect(button.attributes("disabled")).toBeDefined();
    expect(button.text()).toBe("payment.saleUnavailable");
  });

  it("hides remaining count when daily sale plan is available", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 6,
          group_id: 10,
          group_platform: "openai",
          name: "Flash Sale",
          description: "",
          price: 10,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          for_sale: true,
          daily_purchase_limit: 3,
          daily_purchase_remaining: 2,
          daily_sale_starts_at: "09:00",
          daily_sale_ends_at: "18:00",
          daily_sale_status: "available",
          daily_sale_countdown_seconds: 3661,
          daily_sale_available_for_payment: true,
          sort_order: 1,
        },
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("payment.planCard.dailySaleTime");
    expect(wrapper.text()).not.toContain("payment.planCard.availableTodayLabel");
    expect(wrapper.text()).not.toContain("Remaining today");
    expect(wrapper.get("button").attributes("disabled")).toBeUndefined();
  });

  it("shows subscription cycle total quota", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 5,
          group_id: 10,
          group_platform: "openai",
          name: "Annual Early Bird",
          description: "",
          price: 99,
          features: [],
          rate_multiplier: 1,
          validity_days: 365,
          validity_unit: "day",
          for_sale: true,
          daily_purchase_limit: 0,
          subscription_total_limit_usd: 1560,
          sort_order: 1,
        },
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("payment.planCard.perPurchaseCycleQuota");
    expect(wrapper.text()).toContain("$1560");
  });

  it("disables purchase once plan while current subscription is active", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 7,
          group_id: 10,
          group_platform: "openai",
          name: "Once Plan",
          description: "",
          price: 10,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          for_sale: true,
          daily_purchase_limit: 0,
          purchase_once_per_active_subscription: true,
          purchase_once_available_for_payment: false,
          purchase_once_unavailable_until: "2026-06-02T00:00:00Z",
          sort_order: 1,
        },
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("payment.purchaseOnceUnavailable");
    const button = wrapper.get("button");
    expect(button.attributes("disabled")).toBeDefined();
    expect(button.text()).toBe("payment.saleUnavailable");
  });

  it("disables weekly off-day plans and shows sale days", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 8,
          group_id: 10,
          group_platform: "openai",
          name: "Weekly Plan",
          description: "",
          price: 10,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          for_sale: true,
          daily_purchase_limit: 0,
          weekly_sale_days: [1, 3, 5],
          weekly_sale_status: "off_day",
          weekly_sale_available_for_payment: false,
          sort_order: 1,
        },
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("payment.planCard.weeklySaleDays");
    expect(wrapper.text()).toContain("payment.weekdays.mon");
    expect(wrapper.text()).toContain("payment.weekdays.wed");
    expect(wrapper.text()).toContain("payment.weekdays.fri");
    expect(wrapper.text()).toContain("payment.planCard.weeklySaleOffDay");
    const button = wrapper.get("button");
    expect(button.attributes("disabled")).toBeDefined();
    expect(button.text()).toBe("payment.saleUnavailable");
  });

  it("uses weekly off-day text instead of daily sale countdown when both rules apply", () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 9,
          group_id: 10,
          group_platform: "openai",
          name: "Weekly Flash Sale",
          description: "",
          price: 10,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          for_sale: true,
          daily_purchase_limit: 0,
          daily_sale_starts_at: "09:00",
          daily_sale_ends_at: "10:00",
          daily_sale_status: "available",
          daily_sale_countdown_seconds: 2140,
          daily_sale_available_for_payment: false,
          weekly_sale_days: [1, 3, 5],
          weekly_sale_status: "off_day",
          weekly_sale_available_for_payment: false,
          sort_order: 1,
        },
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("payment.planCard.dailySaleTime");
    expect(wrapper.text()).toContain("09:00 - 10:00");
    expect(wrapper.text()).toContain("payment.planCard.weeklySaleOffDay");
    expect(wrapper.text()).not.toContain("payment.planCard.availableNow");
    expect(wrapper.text()).not.toContain("payment.planCard.dailySaleCountdownToEnd");
    const button = wrapper.get("button");
    expect(button.attributes("disabled")).toBeDefined();
  });
});

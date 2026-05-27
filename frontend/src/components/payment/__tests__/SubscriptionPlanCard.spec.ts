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
          soldOutToday: "Sold out today",
          unavailableNow: "Not available yet",
          quota: "Quota",
          rate: "Rate",
          unlimited: "Unlimited",
        },
        saleUnavailable: "Not available",
        subscribeNow: "Subscribe now",
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

  it("shows daily sale window, countdown, and disables unavailable plan", () => {
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
    expect(wrapper.text()).toContain("payment.planCard.availableToday");
    const button = wrapper.get("button");
    expect(button.attributes("disabled")).toBeDefined();
    expect(button.text()).toBe("payment.saleUnavailable");
  });
});

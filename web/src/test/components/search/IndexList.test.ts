import { mount } from "@vue/test-utils";
import { Quasar, Dialog, Notify } from "quasar";
import { expect, it } from "vitest";

import IndexList from "../../../components/search/IndexList.vue";
import store from "../../../store";
import i18n from "../../../locales";

it("should mount IndexList component", async () => {
  const wrapper = mount(IndexList, {
    shallow: false,
    components: {
      Notify,
      Dialog,
    },
    global: {
      plugins: [Quasar, i18n, store],
    },
  });
  expect(IndexList).toBeTruthy();

  // console.log("IndexList is: ", wrapper.html());

  // expect(wrapper.text()).toContain("IndexList");
});

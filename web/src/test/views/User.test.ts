import User from "../../views/User.vue";
import store from "../../store";

import { it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";

import i18n from "../../locales";
import AddUpdateUser from "../../components/user/AddUpdateUser.vue";

it("mount User", async () => {
  const wrapper = mount(User, {
    shallow: true,
    components: {
      AddUpdateUser,
    },
    global: {
      plugins: [Quasar, i18n, store],
    },
  });
  expect(User).toBeTruthy();
  // const wrapper = wrapperFactory();

  console.log("User is", wrapper.html());
});

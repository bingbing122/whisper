<script>
import LoadingPart from "@/components/LoadingPart.vue";
import axios from "axios";

export default {
  components: {
    LoadingPart
  },
  emits: ['skipSearch'],
  props: {
    equipResult: {
      show: 0,
      isLoading: false,
      data: {},
    },
  },
  data() {
    return {
      sideDrawer: {
        isLoading:false,
        show: false,
        title: '合成路线',
        heroes:[],
        loadingHeroes:false,
      },
    }
  },
  watch: {
    'equipResult.show'(show) {
      this.sideDrawer.show = show !== 0;
    },
    'equipResult.isLoading'(isLoading) {
      this.sideDrawer.isLoading = isLoading
    },
    'equipResult.data'(data){
      this.querySuitHeroes(data.current.platform,data.current.version,data.current.ID)
    }
  },
  methods: {
    querySuitHeroes(platform,version,id) {
      this.sideDrawer.loadingHeroes  = true
      axios.post('/equip/hero/suit', {
            'platform': platform,
            'version': version,
            'id': id,
          }
      ).then(response => {
            this.sideDrawer.heroes = response.data.data
          }
      ).catch(error => {
            console.error('Error fetching server data:', error);
          }
      ).finally(() => {
            this.sideDrawer.loadingHeroes = false
          }
      );
    }
  },
  mounted() {

  },
}
</script>

<template>
  <a-drawer
      v-model:open="sideDrawer.show"
      class="custom-class"
      root-class-name="root-class-name"
      :title="sideDrawer.title"
      placement="right"
  >
    <template v-if="sideDrawer.isLoading">
      <a-skeleton active/>
    </template>

    <template v-if="!sideDrawer.isLoading">
      <!-- 合成路线路 START-->
      <div class="equip-roadmap equip-into">
        <!-- 可以合成为 START-->
        <template v-for="(equip ,index) in equipResult.data['into']" :key="index">
          <a-popover placement="bottom" arrow-point-at-center>
            <template #content>
              <div class="roadmap-item">
                <span class="roadmap-item-title">
                  {{ equip.name }}
                </span>
                <a-tag>
                  <span class="roadmap-item-price">
                    价格:{{ equip.price }}
                  </span>
                </a-tag>
                <a-tag>
                  <span class="roadmap-item-price">
                    {{ equip.maps }}
                  </span>
                </a-tag>
              </div>
              <span v-html="equip.desc"></span>
            </template>
            <img @click="$emit('skipSearch',equip.name)"
                 :src="equip.icon">
          </a-popover>
        </template>
        <!-- 可以合成为 END-->

        <!-- 当前装备 START-->
        <a-divider>
          <a-popover placement="bottom" arrow-point-at-center>
            <template #content>
              <div class="roadmap-item">
                <span class="roadmap-item-title">
                  {{ equipResult.data['current'].name }}
                </span>
                <a-tag>
                  <span class="roadmap-item-price">
                    价格:{{ equipResult.data['current'].price }}
                  </span>
                </a-tag>
                <a-tag>
                  <span class="roadmap-item-price">
                    {{ equipResult.data['current'].maps }}
                  </span>
                </a-tag>
              </div>
            </template>
            <img class="equip-roadmap equip-current"
                 :src="equipResult.data['current'].icon"
                 alt="">
          </a-popover>
        </a-divider>
        <!-- 当前装备 END-->

        <!-- 合成自 START-->
        <template v-for="(equip ,index) in equipResult.data['from']" :key="index">
          <a-popover placement="bottom" arrow-point-at-center>
            <template #content>
              <div class="roadmap-item">
                <span class="roadmap-item-title">
                  {{ equip.name }}
                </span>
                <a-tag>
                  <span class="roadmap-item-price">
                    价格:{{ equip.price }}
                  </span>
                </a-tag>
                <a-tag>
                  <span class="roadmap-item-price">
                    {{ equip.maps }}
                  </span>
                </a-tag>
              </div>
              <span v-html="equip.desc"></span>
            </template>
            <img @click="$emit('skipSearch',equip.name)"
                 :src="equip.icon">
          </a-popover>
        </template>
        <!-- 合成自 END-->
      </div>
      <!-- 合成路线路 END-->

      <!-- summary START-->
      <table class="roadmap-detail">
        <tr v-for="(equip ,index) in equipResult.data['from']" :key="index">
          <td>
            <img :src="equip.icon" alt="">
            <span>{{ equip.name }}</span>
            <span>, 价格: <em>{{ equip.price }}</em> </span>
          </td>
        </tr>
        <tr>
          <td>
            总价: <em>{{ equipResult.data['current'].price }}</em> , 合成所需: <em>{{
              equipResult.data.gapPriceFrom
            }}</em>
          </td>
        </tr>
      </table>
      <!-- summary END-->

      <!-- 适配英雄 START-->
      <h4 class="equip-suit-hero" v-if="sideDrawer.loadingHeroes" >适配英雄</h4>
      <LoadingPart v-if="sideDrawer.loadingHeroes" />
      <template v-if="!sideDrawer.loadingHeroes && sideDrawer.heroes.length > 0" >
        <h4 class="equip-suit-hero">适配英雄</h4>
        <template v-for="(hero,i) in sideDrawer.heroes" :key="i">
          <a-tooltip :title="hero.name">
            <img class="equip-hero-avatar"
                 :title="hero.name"
                 :src="hero.icon"
                 :alt="hero.name"
            />
          </a-tooltip>
        </template>
      </template>

      <!-- 适配英雄 END-->
    </template>
  </a-drawer>
</template>

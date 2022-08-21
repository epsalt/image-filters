<script setup lang="ts">
import { reactive, computed } from 'vue'
import ballsRender from './assets/BallsRender.png'

interface State {
  filter: string;
  display: string;
  uploaded: string | ArrayBuffer | null;
  filtered: string | null;
}

interface Response {
  data: string;
}

const filters = [
  {
    endpoint: "greyscale",
    label: "Just Greyscale",
  },
  {
    endpoint: "dither",
    label: "Floyd-Steinberg Dithering",
  },
  {
    endpoint: "blur",
    label: "Gaussian Blur",
  },
]
const state: State = reactive({
  filter: "greyscale",
  uploaded: null,
  filtered: "",
  display: "",
})

fetch(ballsRender)
  .then((data) => data.blob())
  .then((blob) => {
    b46ify(blob).then((data) => {
      if (!data) return
      state.uploaded = btoa(data)
      update();
    })
  })


function b46ify(blob: Blob): Promise<string | undefined> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsBinaryString(blob)
    reader.onload = () => {
      resolve(reader.result?.toString())
    }
    reader.onerror = reject;
  })
}

function handleUpload(event: Event): void {
  const target = event.target as HTMLInputElement;
  if (!target.files || target.files.length < 1) return;
  const file = target.files[0];

  b46ify(file).then((data) => {
    if (!data) return
    state.uploaded = btoa(data)
    update();
  })
}

function update(): void {
  if (!state.uploaded) return

  request<Response>(`/api/${state.filter}`, { data: state.uploaded })
    .then(({ data }) => {
      state.filtered = data
    })
}

function request<Type>(url: string, data: Object): Promise<Type> {
  return fetch(url, {
    method: "POST",
    body: JSON.stringify(data)
  }).then((response) => {
    if (!response.ok) {
      throw new Error(response.statusText)
    }
    return response.json()
  })
}

const preview = computed<string>(() => {
  if (!state.uploaded) return ""
  return `data:image/png;base64,${state.uploaded}`
})

const result = computed<string>(() => {
  if (!state.filtered) return ""
  return `data:image/png;base64,${state.filtered}`
})
</script>

<template>
  <h1>Image filters</h1>
  <p>
    These are some simple image filters implemented in Go. Pick a
    filter and upload an image to try them out. Render of spheres by Wikipedia user Mimigu.
  </p>
  <div v-for="filter in filters">
    <input @change="update" type="radio" :id="filter.endpoint" :value="filter.endpoint" v-model="state.filter" />
    <label :for="filter.endpoint">{{ filter.label }}</label>
  </div>
  <br>
  <form>
    <label for="img">Select image: </label>
    <input @change="handleUpload" type="file" id="img" name="img" accept="image/png">
  </form>
  <img :src="preview">
  <img :src="result">
</template>
<style>
img {
  margin: 10px;
  max-width: 300px;
  max-height: 300px;
}

body {
  margin: 40px auto;
  font-family: sans-serif;
  max-width: 650px;
  line-height: 1.6;
  font-size: 15px;
  color: #444;
  padding: 0 10px
}

h1 {
  line-height: 1.2
}
</style>

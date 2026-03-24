<template>
  <div class="ascii-logo-wrap">
    <img
      class="ascii-logo ascii-logo--dark"
      src="/logo-dark.png"
      alt="yolobox logo rendered from the block ASCII wordmark"
      decoding="async"
    >
    <img
      class="ascii-logo ascii-logo--light"
      src="/logo-light.png"
      alt="yolobox logo rendered from the block ASCII wordmark"
      decoding="async"
    >
  </div>
</template>

<style>
.ascii-logo-wrap {
  width: min(100%, 30rem);
  margin: 0 0 24px;
}

.ascii-logo {
  display: block;
  width: 100%;
  height: auto;
  object-fit: contain;
}

html.dark .ascii-logo--light {
  display: none;
}

html:not(.dark) .ascii-logo--dark {
  display: none;
}

@media (max-width: 700px) {
  .ascii-logo-wrap {
    margin-bottom: 12px;
  }
}
</style>

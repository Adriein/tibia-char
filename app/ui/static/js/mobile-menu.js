(function () {
  var toggle = document.getElementById("sidebar-toggle");
  var sidebar = document.getElementById("sidebar");
  var overlay = document.getElementById("sidebar-overlay");
  var close = document.getElementById("sidebar-close");

  if (!toggle || !sidebar || !overlay || !close) return;

  function open() {
    sidebar.classList.add("sidebar--open");
    overlay.classList.add("sidebar__overlay--visible");
    document.body.style.overflow = "hidden";
  }

  function closeSidebar() {
    sidebar.classList.remove("sidebar--open");
    overlay.classList.remove("sidebar__overlay--visible");
    document.body.style.overflow = "";
  }

  toggle.addEventListener("click", open);
  close.addEventListener("click", closeSidebar);
  overlay.addEventListener("click", closeSidebar);

  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape") closeSidebar();
  });
})();

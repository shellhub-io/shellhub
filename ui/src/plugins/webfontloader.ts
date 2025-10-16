// webfontloader documentation: https://github.com/typekit/webfontloader
const loadFonts = async () => {
  const webFontLoader = await import("webfontloader");

  webFontLoader.load({
    google: {
      families: ["Roboto:100,300,400,500,700,900&display=swap"],
    },
  });
};

export default loadFonts;

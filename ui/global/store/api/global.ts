const premiumContent = async () => {
  const response = await fetch("https://static.shellhub.io/premium-features.v1.json");
  const data = await response.json();
  return data;
};

export default premiumContent;

const showTag = (str?: string): boolean => {
  if (!str) return false;
  return str.length > 10;
};

export default showTag;

const formatDeviceSort = (field, isDesc) => {
  let formatedField = null;
  let formatedStatus = false;
  let ascOrDesc = 'asc';

  if (field !== undefined) {
    formatedField = field === 'hostname' ? 'name' : field; // customize to api field
  }

  if (isDesc !== undefined) {
    formatedStatus = isDesc;
  }

  if (formatedStatus === true) {
    ascOrDesc = 'desc';
  }

  return {
    field: formatedField,
    status: formatedStatus,
    statusString: ascOrDesc,
  };
};

export { formatDeviceSort as default };

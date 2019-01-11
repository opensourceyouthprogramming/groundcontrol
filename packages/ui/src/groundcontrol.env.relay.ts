import {
  Environment,
  Network,
  RecordSource,
  RequestNode,
  Store,
  Variables,
} from "relay-runtime";

const fetchQuery = async (
  operation: RequestNode,
  variables: Variables,
) => {
  return fetch("http://localhost:4000/graphql", {
    body: JSON.stringify({
      query: operation.text,
      variables,
    }),
    headers: {
      "Content-Type": "application/json",
    },
    method: "POST",
  }).then((response) => {
    return response.json();
  });
};

export default new Environment({
  network: Network.create(fetchQuery),
  store: new Store(new RecordSource()),
});
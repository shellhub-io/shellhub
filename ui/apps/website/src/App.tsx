import { Routes, Route } from "react-router-dom";
import { ScrollToTop } from "@/components";
import { routes } from "@/routes";

export default function App() {
  return (
    <>
      <ScrollToTop />
      <Routes>
        {routes.map((r) => (
          <Route key={r.path} path={r.path} element={r.element} />
        ))}
      </Routes>
    </>
  );
}

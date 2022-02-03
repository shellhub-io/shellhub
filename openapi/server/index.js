import koa from "koa"
import koaStatic from "koa-static"

const app = new koa()

app.use(koaStatic("./www"))

app.listen("8080")
console.info("Listing on 8080")
(ns jp.nijohando.naruko.slack-notifier.core
  (:require-macros [cljs.core.async.macros :refer [go]]
                   [cljs.spec.alpha :as s])
  (:require [cljs.nodejs :as node]
            [cljs.spec.alpha :as s]
            [cljs.core.async :refer [promise-chan >!]]
            [cljs.core.async.impl.protocols :refer [ReadPort]]
            [jp.nijohando.failable :refer [fail]]
            [httpurr.client :as http]
            [httpurr.client.node :refer [client]]
            [promesa.core :as p]))

(defn- decode
  [response]
  (update response :body #(js->clj (js/JSON.parse %) :keywordize-keys true)))

(defn- encode
  [request]
  (update request :body #(js/JSON.stringify (clj->js %))))

(s/def ::token string?)
(s/def ::channel string?)
(s/def ::username string?)
(s/def ::msg string?)
(s/def ::request (s/keys :req-un [::token ::channel ::username ::msg]))
(s/fdef notify
  :args (s/cat :request ::request)
  :ret #(satisfies? ReadPort %))
(defn notify [{:keys [token channel username msg] :as rquest}]
  (let [ch (promise-chan)]
    (-> (http/send! client
                    (encode {:url "https://slack.com/api/chat.postMessage"
                             :method :post
                             :headers {"Content-Type" "application/json; charset=utf-8"
                                       "Authorization" (str "Bearer " token)}
                             :body {:username username
                                    :channel channel
                                    :text msg}}))
        (p/then (fn [response]
                  (let [body (-> response decode :body)
                        success? (:ok body)
                        result (if success?  :ok (fail (:error body)))]
                    (go (>! ch result)))))
        (p/catch (fn [error]
                   (go (>! ch (fail error))))))
    ch))

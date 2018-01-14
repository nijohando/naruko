(ns jp.nijohando.naruko.slack-notifier.lambda
  (:require-macros [cljs.core.async.macros :refer [go]])
  (:require [jp.nijohando.naruko.slack-notifier.core :as core]
            [jp.nijohando.failable :refer [failure?]]
            [cljs.core.async :refer [<!]]
            [cljs.nodejs :as node]))

(def env (.-env node/process))

(defn- execute
  [edn]
  (let [token (.-SLACK_API_TOKEN env)
        channel (.-SLACK_CHANNEL env)
        username (.-SLACK_USERNAME env)
        msg(-> edn
               :current
               clj->js
               (#(.stringify js/JSON % nil "  ")))]
    (core/notify {:token token
                  :channel channel
                  :username username
                  :msg msg})))

(defn handler
  [event context callback]
  (let [threshold (-> (.-DETECTION_THRESHOLD env)
                      (js/parseInt 10))
        edn (js->clj event :keywordize-keys true)
        now (get-in edn [:current :updatedAt])
        last-time (get-in edn [:previous :updatedAt])
        diff (- now last-time)]
      (if (< diff threshold)
        (callback nil "below the threshold")
        (go
          (let [result (<! (execute edn))]
            (if (failure? result)
              (callback @result)
              (callback nil "notified")))))))

(node/enable-util-print!)
(aset js/exports "handler" handler)

import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import admin from './admin'
import misc from './misc'
import extensions from './extensions'
import overrides from './overrides'
import { mergeLocaleAdditions, mergeLocaleOverrides } from '../merge'

const messages = {
  ...landing,
  ...common,
  ...dashboard,
  admin,
  ...misc,
}

export default mergeLocaleOverrides(mergeLocaleAdditions(messages, extensions), overrides)

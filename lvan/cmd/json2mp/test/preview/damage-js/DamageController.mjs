
    class DamageController {
        //通用函数==============================================
        get_skill_damage_rate(context) {
            return context.damageRate;
          }
          get_distance(attack_entity, target_entity, user) {
            let atkPos = attack_entity.worldPos;
            let tagPos = target_entity.worldPos;
            let x = atkPos.x - tagPos.x;
            let y = atkPos.y - tagPos.y;
            let z = atkPos.z - tagPos.z;
            let dy = Math.sqrt((x * x) + (y * y) + (z * z));
            return Math.ceil(dy);
          }

          get_skill_slot_count(context, entity, skill_bar_id, used) {
            let count = 0;
            if (entity.dicSkillBar && entity.dicSkillBar[skill_bar_id]) {
              let skillbar = entity.dicSkillBar[skill_bar_id];
              if (entity.type == 1) {
                // 玩家
                // { [barId: number]: pb.SkillBar }
                let info = skillbar;
                let slots = info.slots;
                let len = slots ? slots.length : 0;
                for (let i = 0; i < len; ++i) {
                  let slot = slots[i];
                  let skillId = slot.skillId;
                  if (!used && skillId == 0) {
                    count++;
                  }
                  else if (used && skillId != 0) {
                    count++;
                  }
                }
              }
              if (entity.type == 2) {
                // 怪物
              }
              else {
                //宠物
              }
            }
            return count;
          }
          get_skill_damage_ext(context) {
            return context.damageExt;
          }
          get_attr_value(entity, attId) {
            if (entity && entity.getProp(attId)) {
              return +entity.getProp(attId);
            }
            return 0;
          }
          output_log(context, strLog) {
            if (context.aryLog && strLog) {
              context.aryLog.push(strLog);
            }
          }
          randNum(min, max) {
            return Math.floor(Math.random() * (max - min + 1)) + min
          }
          add_combat_damage(context, damage, combatType) {
            context.damage[combatType] = damage;
          }
          get_entity_type(entity) {
            if (entity) {
              return entity.type;
            }
            return 0;
          }
          get_monster_type(entity) {
            if (entity && entity.type == 2) {
              return entity.cfgType;
            }
            return 0;
          }
          get_entity_buff_count(context, entity, buffLabel) {
            return entity.getBuffCountByTag(buffLabel);
          }
          get_attack_target_count(context) {
            return 1;
          }
          set_final_combat_type(context, combatType) {
            context.combatType = combatType;
          }
          damage_hit(context, rateId, attack_entity, target_entity, user) {
            let value = this["get_hit_rate_" + rateId](attack_entity, target_entity, user);
            if (Math.random() * 10000 <= value) {
              return true;
            }
            return false;
          }

          add_fight_info(context, fightId, suc, v){
            context.addFightInfo(fightId,suc,v)
          }

          get_buff_lefttime(entity, buffid){
            if(entity)
                return entity.getBuffLeftTime(buffid);
            return 0;
          }

          get_buff_count(entity, buffid){
              if(entity)
                  return entity.getBuffCount(buffid);
              return 0;
          }

        get_buff_interval(entity, buffid){
            if(entity)
                return entity.getBuffInterval(buffid);
            return 0;
        }
        //伤害公式子函数=========================================
        //伤害公式函数===========================================
        
        get_hit_rate_10001(context, attack_entity, target_entity, user) {
  return   10000
  }
  
damage_10001(context, init_damage, attack_entity, target_entity, user) {
  init_damage =   0
  if (this.damage_hit(context, 10001, attack_entity, target_entity, user)) { 
    this.add_combat_damage(context,   100, 10001)
  
  } else {
  
  }
  return   0
  }
  

        
        get_hit_rate_10002(context, attack_entity, target_entity, user) {
  return   Math.max( 0, ( 10000 - this.get_attr_value(target_entity,   10065) ) )
  }
  
damage_10002(context, init_damage, attack_entity, current_entity, target_entity, user) {
  init_damage =   this.get_attr_value(attack_entity,   10007)
  if (this.damage_hit(context, 10002, attack_entity, target_entity, user)) { 
    // 技能基础伤害
    var BaseFlow001 =   Math.max( 1, ( Math.floor( ( init_damage * context.getCustomValue('par001') ) / 10000 ) + context.getCustomValue('par002') ) )
    this.add_combat_damage(context,   BaseFlow001, 10001)
  
  } else {
    var BaseFlow001 =   0
    this.add_combat_damage(context,   BaseFlow001, 10001)
  
  }
  return   BaseFlow001
  }
  

        
        get_hit_rate_10003(context, attack_entity, target_entity, user) {
  return   10000
  }
  
damage_10003(context, init_damage, attack_entity, current_entity, target_entity, user) {
  init_damage =   this.get_attr_value(attack_entity,   10007)
  if (this.damage_hit(context, 10003, attack_entity, target_entity, user)) { 
    var zl001 =   Math.ceil( ( init_damage * this.get_skill_damage_rate(context) ) / 10000 )
    this.add_combat_damage(context,   zl001, 10001)
  
  } else {
    var zl001 =   0
    this.add_combat_damage(context,   zl001, 10001)
  
  }
  return   zl001
  }
  

        
    }

    class BridgeUtil {
        /**获取伤害公式函数 */
        static getDamageFormulaById(id, context, init_damage, attack_entity, target_entity, user) {
          return this._controller[BridgeUtil.prefix + id].call(this._controller, context, init_damage, attack_entity, target_entity, user);
        }
      }
      BridgeUtil._controller = new DamageController();
      BridgeUtil.prefix = "damage_";

export{BridgeUtil}